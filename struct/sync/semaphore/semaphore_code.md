##semaphore源码分析以及总结
    go中的semaphore，提供sleep和wakeup原语，使其能够在其它同步原语中的竞争情况下使用。
    一个goroutine需要休眠时，将其进行集中存放，当需要wakeup时，再将其取出，重新放入调度器中。

    比如在读写锁实现中，读写锁之间相互阻塞唤醒，就是通过sleep和wakeup实现，当读锁存在时，新加入的写锁通过sema
    阻塞自己，当前面的读锁完成，再通过sema唤醒被阻塞的写锁。
    semaphore的实现使用到了sudog。sudog是预先是用来存放处于阻塞状态的goroutine的一个上层抽象，
    是用来实现用户态信号量的主要机制之一。
    例如当一个goroutine因为等待channel的数据需要进行阻塞时，sudog会将goroutine及用于等待数据的位置进行记录，
    并进而串联成一个等待队列，或者二叉平衡树
    
    Acquire方法阻塞的获取指定权重的资源，如果没有空闲的资源，会进去休眠等待。
    TryAcquire方法非阻塞的获取指定权重的资源，如果当前没有空闲资源，会直接返回false
    TryAcquire获取权重为n的信号量而不阻塞，相比Acquire少了等待队列的处理。
    Release用于释放指定权重的资源，如果有waiters则尝试去一一唤醒waiter。唤醒的时候先进先出，避免资源大的waiter被饿死

    Acquire和 TryAcquire方法都可以用于获取资源，前者会阻塞地获取信号量。后者会非阻塞地获取信号量，如果获取不到就返回false。
    Release归还信号量后，会以先进先出的顺序唤醒等待队列中的调用者。如果现有资源不够处于等待队列前面的调用者请求的资源数，所有等待者会继续等待。
    如果一个goroutine申请较多的资源，由于上面说的归还后唤醒等待者的策略，它可能会等待比较长的时间。



参考文档：
- [go中x/sync/semaphore解读](https://www.cnblogs.com/ricklz/p/14604678.html)
- [Go并发编程实战--信号量的使用方法和其实现原理](https://juejin.cn/post/6906677772479889422)
- [Golang并发同步原语之-信号量Semaphore](https://blog.haohtml.com/archives/25563)
- [Go并发编程实战–信号量的使用方法和其实现原理](https://juejin.cn/post/6906677772479889422)

