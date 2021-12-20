
writer 加解锁过程:

```go
// Lock 对 rw 加写锁
// 如果当前锁已经被锁住进行读或者进行写
// Lock 会阻塞，直到锁可用
func (rw *RWMutex) Lock() {
    // First, resolve competition with other writers.
    // 首先需要解决和其它 writer 进行的竞争，这里是去抢 RWMutex 中的 Mutex 锁
    rw.w.Lock()
    // 抢到了上面的锁之后，通知所有 reader，现在有一个挂起的 writer 等待写入了
    r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
    // 等待最后的 reader 将其唤醒
    if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
        runtime_Semacquire(&rw.writerSem)
    }
}

// Unlock 将 rw 的读锁解锁。如果当前 rw 没有处于锁定读的状态，那么就是 bug
//
// 像 Mutex 一样，一个上锁的 RWMutex 并没有和特定的 goroutine 绑定。
// 可以由一个 goroutine Lock 它，并由其它的 goroutine 解锁
func (rw *RWMutex) Unlock() {

    // 告诉所有 reader 现在没有活跃的 writer 了
    r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)
    if r >= rwmutexMaxReaders {
        throw("sync: Unlock of unlocked RWMutex")
    }
    // Unblock 掉所有正在阻塞的 reader
    for i := 0; i < int(r); i++ {
        runtime_Semrelease(&rw.readerSem, false)
    }
    // 让其它的 writer 可以继续工作
    rw.w.Unlock()
}
