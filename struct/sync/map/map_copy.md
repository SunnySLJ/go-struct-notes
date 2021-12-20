## sync.Ma
数据结构:

```go

// sync.Map 类似 map[interface{}]interface{}，但是是并发安全的。
// Load，Store 和 delete 是通过延迟操作来均摊成本而达到常数时间返回的
//
// sync.Map 类型是为特殊目的准备的，一般的代码还是应该使用普通的 Go map 类型，
// 并自己完成加锁和多线程协作，这样能够有更好的类型安全并且更易维护。
//
// 这里的 Map 是为了两种用例来优化的:
// (1) 当某个指定的 key 只会被写入一次，但是会被读取非常多次，例如像不断增长的 caches。
// (2) 当多个 goroutine 分别分布读、写和覆盖不同的 key。
// 这两种场景下，使用 sync.Map，相比普通的 map 配合 Mutex 或 RWMutex，可以大大降低锁的竞争
//
// Map 的零值是可以使用的空 Map。当 Map 被首次使用之后，就不能再被拷贝了
type Map struct {
    mu Mutex

    // read 包含了 map 内容的一部分，这部分进行并发访问是安全的(无论持有 mu 或未持有)
    //
    // read 字段本身对 load 操作而言是永远安全的，但执行 store 操作时，必须持有 mu
    //
    // 在 read 中存储的 entries 可以在不持有 mu 的情况下进行并发更新，但更新一个之前被 expunged
    // 的项需要持有 mu 的前提下，将该 entry 被拷贝到 dirty map 并将其 unexpunged(就是 expunged 的逆操作)
    read atomic.Value // readOnly

    // dirty 包含了 map 中那些在访问时需要持有 mu 的部分内容
    // 为了确保 dirty map 的元素能够被快速地移动到 read map 中
    // 它也包含了那些 read map 中未删除(non-expunged)的项
    //
    // expunged 掉的 entries 不会在 dirty map 中存储。被 expunged 的 entry，
    // 如果要存新值，需要先执行 expunged 逆操作，然后添加到 dirty map，然后再进行更新
    //
    // 如果 dirty map 为 nil，下一个对 map 的写入会初始化该 map，并对干净 map 进行一次浅拷贝
    // 并忽略那些过期的 entry
    dirty map[interface{}]*entry

    // misses 计算从 read map 上一次被更新开始的需要 lock mu 来进行的 load 次数
    //
    // 一旦发生了足够多的 misses 次数，足以覆盖到拷贝 dirty map 的成本，dirty map 就会被合并进
    // read map(在 unamended 状态下)，并且下一次的 store 操作则会生成一个新的 dirty map
    misses int
}

// readOnly 是原子地存在 Map.read 字段中的不可变结构
type readOnly struct {
    m       map[interface{}]*entry
    amended bool // 如果 dirty map 中包含有不在 m 中的项，那么 amended = true
}

// expunged 是一个任意类型的指针，用来标记从 dirty map 中删除的项
var expunged = unsafe.Pointer(new(interface{}))

// entry 是 map 对应一个特定 key 的槽
type entry struct {
    // p 指向 entry 对应的 interface{} 类型的 value
    //
    // 如果 p == nil，那么说明对应的 entry 被删除了，曲 m.dirty == nil
    //
    // 如果 p == expunged，说明 entry 被删除了，但 m.dirty != nil，且该 entry 在 m.dirty 中不存在
    //
    // 除了上述两种情况之外，entry 则是合法的值并且在 m.read.m[key] 中存在
    // 如果 m.dirty != nil，也会在 m.dirty[key] 中
    //
    // 一个 entry 项可以被 atomic cas 替换为 nil 来进行删除: 当 m.dirty 之后被创建的话，
    // 会原子地将 nil 替换为 expunged，且不设置 m.dirty[key] 的值。
    //
    // 一个 entry 对应的值可以用 atomic cas 来更新，前提是 p != expunged。
    // 如果 p == expunged，entry 对应的值只能在首次赋值 m.dirty[key] = e 之后进行
    // 这样查找操作可以用 dirty map 来找到这个 entry
    p unsafe.Pointer // *interface{}
}

```

Load:

```go
func newEntry(i interface{}) *entry {
    return &entry{p: unsafe.Pointer(&i)}
}

// 返回 map 中 key 对应的值，如果不存在，返回 nil
// ok 会返回该值是否在 map 中存在
func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
    read, _ := m.read.Load().(readOnly)
    e, ok := read.m[key]
    if !ok && read.amended {
        m.mu.Lock()
        // 当 m.dirty 已被合并到 read 阻塞在 m.mu 时，避免报告不应该报告的 miss (如果未来的
        // 同一个 key 的 loads 不会发生 miss，那么为了这个 key 而拷贝 dirty map 就不值得了)
        read, _ = m.read.Load().(readOnly)
        e, ok = read.m[key]
        if !ok && read.amended {
            e, ok = m.dirty[key]
            // 无论 entry 是否存在，都记录一次 miss:
            // 这个 key 会始终走 slow path(加锁的)，直到 dirty map 被合并到 read map
            m.missLocked()
        }
        m.mu.Unlock()
    }
    if !ok {
        return nil, false
    }
    return e.load()
}

func (e *entry) load() (value interface{}, ok bool) {
    p := atomic.LoadPointer(&e.p)
    if p == nil || p == expunged {
        return nil, false
    }
    return *(*interface{})(p), true
}

```

Store:

```go

// 为某一个 key 的 value 赋值
func (m *Map) Store(key, value interface{}) {
    read, _ := m.read.Load().(readOnly)
    if e, ok := read.m[key]; ok && e.tryStore(&value) {
        return
    }

    m.mu.Lock()
    read, _ = m.read.Load().(readOnly)
    if e, ok := read.m[key]; ok {
        if e.unexpungeLocked() {
            // entry 之前被 expunged 了，表明这时候存在非空的 dirty map
            // 且该 entry 不在其中
            m.dirty[key] = e
        }
        e.storeLocked(&value)
    } else if e, ok := m.dirty[key]; ok {
        e.storeLocked(&value)
    } else {
        if !read.amended {
            // 为 dirty map 增加第一个新的 key
            // 确保分配内存，并标记 read-only map 为 incomplete(amended = true)
            m.dirtyLocked()
            m.read.Store(readOnly{m: read.m, amended: true})
        }
        m.dirty[key] = newEntry(value)
    }
    m.mu.Unlock()
}

// tryStore 在 entry 没有被 expunged 时存储 value
//
// 如果 entry 被 expunged 了，tryStore 会返回 false 并且放弃对 entry 的 value 赋值
func (e *entry) tryStore(i *interface{}) bool {
    p := atomic.LoadPointer(&e.p)
    if p == expunged {
        return false
    }
    for {
        if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
            return true
        }
        p = atomic.LoadPointer(&e.p)
        if p == expunged {
            return false
        }
    }
}

// unexpungeLocked 确保该 entry 没有被标记为 expunged
//
// 如果 entry 之前被 expunged 了，它必须在 m.mu 被解锁前，被添加到 dirty map
func (e *entry) unexpungeLocked() (wasExpunged bool) {
    return atomic.CompareAndSwapPointer(&e.p, expunged, nil)
}

// storeLocked 无条件地存储对应 entry 的值
//
// entry 必须未被 expunged
func (e *entry) storeLocked(i *interface{}) {
    atomic.StorePointer(&e.p, unsafe.Pointer(i))
}

```

LoadOrStore:

```go
// LoadOrStore 如果 key 对应的值存在，那么就返回
// 否则的话会将参数的 value 存储起来，并返回该值
// loaded 如果为 true，表示值是被加载的，false 则表示实际执行的是 store 操作
func (m *Map) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
    // 确定命中的话，避免锁
    read, _ := m.read.Load().(readOnly)
    if e, ok := read.m[key]; ok {
        actual, loaded, ok := e.tryLoadOrStore(value)
        if ok {
            return actual, loaded
        }
    }

    m.mu.Lock()
    read, _ = m.read.Load().(readOnly)
    if e, ok := read.m[key]; ok {
        if e.unexpungeLocked() {
            m.dirty[key] = e
        }
        actual, loaded, _ = e.tryLoadOrStore(value)
    } else if e, ok := m.dirty[key]; ok {
        actual, loaded, _ = e.tryLoadOrStore(value)
        m.missLocked()
    } else {
        if !read.amended {
            // 为 dirty map 添加第一个新 key
            // 确保分配好内存，并标记 read-only map 为 incomplete
            m.dirtyLocked()
            m.read.Store(readOnly{m: read.m, amended: true})
        }
        m.dirty[key] = newEntry(value)
        actual, loaded = value, false
    }
    m.mu.Unlock()

    return actual, loaded
}

// tryLoadOrStore 原子地 load 或 store 一个 entry 对应的 value
// 前提是该 entry 没有被 expunged
//
// 如果 entry 被 expunged 的话，tryLoadOrStore 会不做任何修改，并返回 ok==false
func (e *entry) tryLoadOrStore(i interface{}) (actual interface{}, loaded, ok bool) {
    p := atomic.LoadPointer(&e.p)
    if p == expunged {
        return nil, false, false
    }
    if p != nil {
        return *(*interface{})(p), true, true
    }

    // 首次 load 之后拷贝 interface，以使该方法对逃逸分析更顺从: 如果我们触发了 "load" 路径
    // 或者 entry 被 expunged 了，我们不应该造成 heap-allocating
    ic := i
    for {
        if atomic.CompareAndSwapPointer(&e.p, nil, unsafe.Pointer(&ic)) {
            return i, false, true
        }
        p = atomic.LoadPointer(&e.p)
        if p == expunged {
            return nil, false, false
        }
        if p != nil {
            return *(*interface{})(p), true, true
        }
    }
}

```

Delete:

```go
// Delete 删除 key 对应的 value
func (m *Map) Delete(key interface{}) {
    read, _ := m.read.Load().(readOnly)
    e, ok := read.m[key]
    if !ok && read.amended {
        m.mu.Lock()
        read, _ = m.read.Load().(readOnly)
        e, ok = read.m[key]
        if !ok && read.amended {
            delete(m.dirty, key)
        }
        m.mu.Unlock()
    }
    if ok {
        e.delete()
    }
}

func (e *entry) delete() (hadValue bool) {
    for {
        p := atomic.LoadPointer(&e.p)
        if p == nil || p == expunged {
            return false
        }
        if atomic.CompareAndSwapPointer(&e.p, p, nil) {
            return true
        }
    }
}

```

Range 遍历:

```go

// Range 按每个 key 和 value 在 map 里出现的顺序调用 f
// 如果 f 返回 false，range 会停止迭代
//
// Range 不需要严格地对应 Map 内容的某个快照: 就是说，每个 key 只会被访问一次，
// 但如果存在某个 key 被并发地更新或者删除了，Range 可以任意地返回修改前或修改后的值
//
// Range 不管 Map 中有多少元素都是 O(N) 的时间复杂度
// 即使 f 在一定数量的调用之后返回 false 也一样
func (m *Map) Range(f func(key, value interface{}) bool) {
    // 需要在一开始调用 Range 的时候，就能够迭代所有已经在 map 里的 key
    // 如果 read.amended 是 false，那么 read.m 就可以满足要求了
    // 而不需要我们长时间持有 m.mu
    read, _ := m.read.Load().(readOnly)
    if read.amended {
        // m.dirty 包含有 未出现在 read.m 中的 key。幸运的是 Range 已经是 O(N) 了
        // (假设 caller 不会中途打断)，所以对于 Range 的调用必然会分阶段完整地拷贝整个 map:
        // 这时候我们可以直接把 dirty map 拷贝到 read!
        m.mu.Lock()
        read, _ = m.read.Load().(readOnly)
        if read.amended {
            read = readOnly{m: m.dirty}
            m.read.Store(read)
            m.dirty = nil
            m.misses = 0
        }
        m.mu.Unlock()
    }

    for k, e := range read.m {
        v, ok := e.load()
        if !ok {
            continue
        }
        if !f(k, v) {
            break
        }
    }
}

```

其它小函数:

```go

func (m *Map) missLocked() {
    m.misses++
    if m.misses < len(m.dirty) {
        return
    }
    m.read.Store(readOnly{m: m.dirty})
    m.dirty = nil
    m.misses = 0
}

func (m *Map) dirtyLocked() {
    if m.dirty != nil {
        return
    }

    read, _ := m.read.Load().(readOnly)
    m.dirty = make(map[interface{}]*entry, len(read.m))
    for k, e := range read.m {
        if !e.tryExpungeLocked() {
            m.dirty[k] = e
        }
    }
}

func (e *entry) tryExpungeLocked() (isExpunged bool) {
    p := atomic.LoadPointer(&e.p)
    for p == nil {
        if atomic.CompareAndSwapPointer(&e.p, nil, expunged) {
            return true
        }
        p = atomic.LoadPointer(&e.p)
    }
    return p == expunged
}
```
- sync.Map 利用了读写分离的思路为读多写少或读写不同 key 的场景而设计，
- 
- 当违背这种设计初衷来使用 sync.Map 的时候性能或许达不到你的期待
- 可以参考下其他诸如散列思路减少锁开销的并发安全
- [Map](https://github.com/orcaman/concurrent-map/
)
# 参考资料

http://www.weixianmanbu.com/article/736.html

https://www.cnblogs.com/gaochundong/p/lock_free_programming.html

https://en.cppreference.com/w/cpp/atomic/memory_order

https://github.com/brpc/brpc/blob/master/docs/cn/atomic_instructions.md

FAQ

Q: 非饥饿模式下，有可能不一定是先lock的先释放呗？

A: 

Q: 既然被选中了唤醒的那个G，说明这个G就是一定要退出lock方法了，也就是被选中的抢到锁的人了。

A:


<img width="330px"  src="https://xargin.com/content/images/2021/05/wechat.png">
