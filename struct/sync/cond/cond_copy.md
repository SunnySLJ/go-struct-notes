## sync.Cond

sync.Cond 本质上就是利用 Mutex 或 RWMutex 的 Lock 会阻塞，来实现了一套事件通知机制。

```go


// Wait 会原子地解锁 c.L，并挂起当前调用 Wait 的 goroutine
// 之后恢复执行时，Wait 在返回之前对 c.L 加锁。和其它系统不一样
// Wait 在被 Broadcast 或 Signal 唤醒之前，是不能返回的
//
// 因为 c.L 在 Wait 第一次恢复执行之后是没有被锁住的，调用方
// 在 Wait 返回之后没办法假定 condition 为 true。
// 因此，调用方应该在循环中调用 Wait
//
//    c.L.Lock()
//    for !condition() {
//        c.Wait()
//    }
//    .. 这时候 condition 一定为 true..
//    c.L.Unlock()
//
func (c *Cond) Wait() {
    c.checker.check()
    t := runtime_notifyListAdd(&c.notify)
    c.L.Unlock()
    runtime_notifyListWait(&c.notify, t)
    c.L.Lock()
}

// Signal 只唤醒等待在 c 上的一个 goroutine。
// 对于 caller 来说在调用 Signal 时持有 c.L 也是允许的，不过没有必要
func (c *Cond) Signal() {
    c.checker.check()
    runtime_notifyListNotifyOne(&c.notify)
}

// Broadcast 唤醒所有在 c 上等待的 goroutine
// 同样在调用 Broadcast 时，可以持有 c.L，但没必要
func (c *Cond) Broadcast() {
    c.checker.check()
    runtime_notifyListNotifyAll(&c.notify)
}

// 检查结构体是否被拷贝过，因为其持有指向自身的指针
// 指针值和实际地址不一致时，即说明发生了拷贝
type copyChecker uintptr

func (c *copyChecker) check() {
    if uintptr(*c) != uintptr(unsafe.Pointer(c)) &&
        !atomic.CompareAndSwapUintptr((*uintptr)(c), 0, uintptr(unsafe.Pointer(c))) &&
        uintptr(*c) != uintptr(unsafe.Pointer(c)) {
        panic("sync.Cond is copied")
    }
}

// noCopy may be embedded into structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock() {}

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
-  sync.Map 利用了读写分离的思路为读多写少或读写不同 key 的场景而设计，当违背这种设计初衷来使用 sync.Map 的时候性能或许达不到你的期待
-  可以参考下其他诸如散列思路减少锁开销的并发安全 [Map](https://github.com/orcaman/concurrent-map/
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
