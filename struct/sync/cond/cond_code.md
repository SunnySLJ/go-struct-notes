##sync.cond源码分析以及总结
    Cond 本质上就是利用 Mutex 或 RWMutex 的 Lock 会阻塞，来实现了一套事件通知机制

    noCopy可以嵌入到结构中，在第一次使用后不可复制,使用go vet作为检测使用
	L根据需求初始化不同的锁，如*Mutex 和 *RWMutex
	notify通知列表,调用Wait()方法的goroutine会被放入list中,每次唤醒,从这里取出
	checker复制检查,检查cond实例是否被复制

    sync.NewCond(l Locker): 新建一个 sync.Cond 变量。注意该函数需要一个Locker 作为必填参数，这是因为在 cond.Wait() 中底层会涉及到 Locker 的锁操作。
    cond.Wait(): 等待被唤醒。唤醒期间会解锁并切走 goroutine。
    cond.Signal(): 只唤醒一个最先 Wait 的 goroutine。对应的另外一个唤醒函数是 Broadcast，
    区别是 Signal 一次只会唤醒一个 goroutine，而 Broadcast 会将全部 Wait 的 goroutine 都唤醒。

##总结：sync.Cond 条件变量用来协调想要访问共享资源的那些 goroutine，当共享资源的状态发生变化的时候，它可以用来通知被互斥锁阻塞的 goroutine

参考文档：
- [sync.Cond源码分析](https://juejin.cn/post/6844904034483044366)

 


