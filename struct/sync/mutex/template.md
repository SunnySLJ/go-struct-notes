##源码文章模板，这样的格式能够很容易的理解，而且内容也齐全，可以跟着学习一下

##https://juejin.cn/post/6968853664718913543

RWMutex是一个支持并行读串行写的读写锁。RWMutex具有写操作优先的特点，写操作发生时，仅允许正在执行的读操作执行，后续的读操作都会被阻塞。
使用场景
RWMutex常用于大量并发读，少量并发写的场景；比如微服务配置更新、交易路由缓存等场景。相对于Mutex互斥锁，RWMutex读写锁具有更好的读性能。
下面以 “多个协程并行读取str变量，一个协程每100毫秒定时更新str变量” 场景为例，进行RWMutex读写锁和Mutex互斥锁的性能对比。
// 基于RWMutex的实现
var rwLock sync.RWMutex
var str1 = "hello"

func readWithRWLock() string {
rwLock.RLock()
defer rwLock.RUnlock()
return str1
}

func writeWithRWLock() {
rwLock.Lock()
str1 = time.Now().Format("20060102150405")
rwLock.Unlock()
}

// 多个协程并行读取string变量，同时每100ms对string变量进行1次更新
func BenchmarkRWMutex(b *testing.B) {
ticker := time.NewTicker(100 * time.Millisecond)
go func() {
for range ticker.C {
writeWithRWLock()
}
}()
b.ResetTimer()
b.RunParallel(func(pb *testing.PB) {
for pb.Next() {
readWithRWLock()
}
})
}
// 基于Mutex实现
var lock sync.Mutex
var str2 = "hello"

func readWithMutex() string {
lock.Lock()
defer lock.Unlock()
return str2
}

func writeWithMutex() {
lock.Lock()
str2 = time.Now().Format("20060102150405")
lock.Unlock()
}

// 多个协程并行读取string变量，同时每100ms对string变量进行1次更新
func BenchmarkMutex(b *testing.B) {
ticker := time.NewTicker(100 * time.Millisecond)
go func() {
for range ticker.C {
writeWithMutex()
}
}()
b.ResetTimer()
b.RunParallel(func(pb *testing.PB) {
for pb.Next() {
readWithMutex()
}
})
}
复制代码
RWMutex读写锁和Mutex互斥锁的性能对比，结果如下：
# go test 结果
go test -bench . -benchtime=10s
BenchmarkRWMutex-8      227611413               49.5 ns/op
BenchmarkMutex-8        135363408               87.8 ns/op
PASS
ok      demo    37.800s
复制代码
源码解析
RWMutex是一个写操作优先的读写锁，如下图所示：

写操作C发生时，读操作A和读操作B正在执行，因此写操作C被挂起；
当读操作D发生时，由于存在写操作C等待锁，所以读操作D被挂起；
读操作A和读操作B执行完成，由于没有读操作和写操作正在执行，写操作C被唤醒执行；
当读操作E发生时，由于写操作C正在执行，所以读操作E被挂起；
当写操作C执行完成后，读操作D和读操作E被唤醒；


RWMutex结构体
RWMutex由如下变量组成：

rwmutexMaxReaders：表示RWMutex能接受的最大读协程数量，超过rwmutexMaxReaders后会发生panic；
w：Mutex互斥锁，用于实现写操作之间的互斥
writerSem：写操作操作信号量；当存在读操作时，写操作会被挂起；读操作全部完成后，通过writerSem信号量唤醒写操作；
readerSem：读操作信号量；当存在写操作时，读操作会被挂起；写操作完成后，通过readerSem信号量唤醒读操作；
readerCount：正在执行中的读操作数量；当不存在写操作时从0开始计数，为正数；当存在写操作时从负的rwmutexMaxReaders开始计数，为负数；
readerWait：写操作等待读操作的数量；当执行Lock()方法时，如果当前存在读操作，会将读操作的数量记录在readerWait中，并挂起写操作；读操作执行完成后，会更新readerWait，当readerWait为0时，唤醒写操作；

const rwmutexMaxReaders = 1 << 30

type RWMutex struct {
w           Mutex  // Mutex互斥锁，用于实现写操作之间的互斥

    writerSem   uint32 // 写操作信号量，用于读操作唤醒写操作
    readerSem   uint32 // 读操作信号量，用于写操作唤醒读操作

    readerCount int32  // 读操作的数量，不存在写操作时从0开始计数，存在写操作时从-rwmutexMaxReaders开始计数
    readerWait  int32  // 写操作等待读操作的数量
}
复制代码
Lock()方法
Lock方法用于写操作获取锁，其操作如下：

获取w互斥锁，保证同一时刻只有一个写操作执行；
将readerCount更新为负数，使后续发生的读操作被阻塞；
如果当前存在活跃的读操作r != 0，写操作进入阻塞状态runtime_SemacquireMutex；

func (rw *RWMutex) Lock() {
// 写操作之间通过w互斥锁实现互斥
rw.w.Lock()
// 1.将readerCount更新为负值，表示当前有写操作；当readerCount为负数时，新的读操作会被挂起
// 2.r表示当前正在执行的读操作数量
r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
// r != 0表示当前存在正在执行的读操作；写操作需要等待所有读操作执行完，才能被执行；
if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
// 将写操作挂起
runtime_SemacquireMutex(&rw.writerSem, false, 0)
}
}
复制代码
Unlock()方法
Unlock方法用于写操作释放锁，其操作如下：

将readerCount更新为正数，表示当前不存在活跃的写操作；

如果更新后的readerCount大于0，表示当前写操作阻塞了readerCount个读操作，需要将所有被阻塞的读操作都唤醒；


将w互斥锁释放，允许其他写操作执行；

func (rw *RWMutex) Unlock() {
// 将readerCount更新为正数，从0开始计数
r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)
if r >= rwmutexMaxReaders {
throw("sync: Unlock of unlocked RWMutex")
}
// 唤醒所有等待写操作的读操作
for i := 0; i < int(r); i++ {
runtime_Semrelease(&rw.readerSem, false, 0)
}
// 释放w互斥锁，允许其他写操作进入
rw.w.Unlock()
}
复制代码
RLock()方法
RLock方法用于读操作获取锁，其操作如下：

原子更新readerCount+1；
如果当前存在写操作atomic.AddInt32(&rw.readerCount, 1) < 0，读操作进入阻塞状态；

func (rw *RWMutex) RLock() {
// 原子更新readerCount+1
// 1. readerCount+1为负数时，表示当前存在写操作；读操作需要等待写操作执行完，才能被执行
// 2. readerCount+1不为负数时，表示当前不存在写操作，读操作可以执行
if atomic.AddInt32(&rw.readerCount, 1) < 0 {
// 将读操作挂起
runtime_SemacquireMutex(&rw.readerSem, false, 0)
}
}
复制代码
RUnlock()方法
RUnlock方法用于读操作释放锁，其操作如下：

原子更新readerCount-1；
如果当前读操作阻塞了写操作atomic.AddInt32(&rw.readerCount, -1)<0，原子更新readerWait-1；

当readerWait为0时，表示阻塞写操作的所有读操作都执行完了，唤醒写操作；



func (rw *RWMutex) RUnlock() {
// 原子更新readerCount-1
// 当readerCount-1为负时，表示当前读操作阻塞了写操作，需要进行readerWait的更新
if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {
rw.rUnlockSlow(r)
}
}

func (rw *RWMutex) rUnlockSlow(r int32) {
if r+1 == 0 || r+1 == -rwmutexMaxReaders {
throw("sync: RUnlock of unlocked RWMutex")
}
// 原子操作readerWait-1
// 当readerWait-1为0时，表示导致写操作阻塞的所有读操作都执行完，将写操作唤醒
if atomic.AddInt32(&rw.readerWait, -1) == 0 {
// 唤醒读操作
runtime_Semrelease(&rw.writerSem, false, 1)
}
}

作者：Shine4YG
链接：https://juejin.cn/post/6968853664718913543
来源：稀土掘金
著作权归作者所有。商业转载请联系作者获得授权，非商业转载请注明出处。
