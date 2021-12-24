##timer源码分析以及总结
    NewTimer 创建一个计时器，在达到指定时间时，将发送当前时间到 Timer.C 通道，调用startTimer将当前timer添加到P的堆中
    AfterFunc 创建一个计时器，但并不会发送值到 Timer.C 通道
    Reset 重置修改定时时间参数 d，如果计时器已过期或已停止，则返回false，否则为true
    如果当前定时器已从P堆中删除，则重新加入P堆中；
    如果修改后的时间提前了，则修改状态为 timerModifiedEarlier，同时唤醒netpool中休眠的线程。
    Stop 停止计时器的运行。如果计时器已过期，则返回false，否则为 true。切记stop并不会关闭channel通道，
    对于timer 的删除不能直接从堆中删除，因为它可能不在当前的P，而是在在其它的P堆上，所以只能将其标记为删除状态，在适当的时候将自动删除
    否则有可能出现其它goroutine向一个已关闭的channel写数据导致的 panic。

     1.timer是一次性定时器、ticker是周期性定时器
     2.timerproc为系统协程的具体实现。
        处理流程：new一个定时器，当触发时间未到，进入睡眠，到触发时间了，开始触发时间，是周期定时器，则重新设置时间并调整堆，若不是，删除timer并调整堆
     3.资源泄露问题，创建ticker的协程不负责结束，只负责从ticker的管道中获取事件，系统协程值负责定时器执行，如果创建了ticker，
        该协程将持续监控ticker的timer，定期触发时间，需要用stop()方法显示关闭，如果未显示关闭，系统协程负担会越来越重，最终将消耗大量的CPU资源
     4.timerproc底层是使用了四叉堆，早起全局只有一个timer heap，所有timer操作抢同一把锁，并发上去的时候影响很大，后面将timer堆拆封成64，
        降低锁的粒度，timer heap分配给不同的GPM的P绑定去执行。根据核数去区分，也有个问题，CPU密集计算任务会导致timer唤醒延迟。
     5.现在的版本timer heap和GPM中的P绑定，去除唤醒的goroutine，timer到期在checkTimers中进行，将timer分给schedule调度循环去进行。
       工作窃取，在调度中work-stealing中用runqsteal会从其他P那里偷timer执行，
     6.runtime.sysmon会为timer未被触发(timeSleepUntil)兜底，启动新线程，启动新线程

    Timer 使用中的坑
    确实 timer 是我们开发中比较常用的工具，但是 timer 也是最容易导致内存泄露，CPU 狂飙的杀手之一。
    不过仔细分析可以发现，其实能够造成问题就两个方面：
    错误创建很多的 timer，导致资源浪费
    由于 Stop 时不会主动关闭 C，导致程序阻塞

    使用 time.Reset 重置 timer，重复利用 timer

## 总结:
    timer的全局堆是一个四叉堆，Goruntine调度timer时触发时间更早的timer,要减少其查询次数尽快被触发。所以四叉树的父节点的触发时间是一定小于子节点的。
    为了兼顾四叉树插、删除、重排速度，所以四个兄弟节点间并不要求其按触发早晚排序。
    现在1.14版本之后timer heap和GMP中的P绑定，timer到期在checkTimers中进行，将timer分给schedule调度循环进行调度。
    其中会进行工作窃取，在work-stealing中会从其他P那里偷timer来执行
    timer是一次性定时器，ticker是周期性的定时器
    NewTimer创建一个计时器，到达指定时间时，将发送当前时间到Timer.C通道，调用startTimer将当前timer添加到P的堆中。
    AfterFunc创建一个计时器，但不会发送值到Timer.C通道
    runtime.sysmon会为timer未被触发(timeSleepUntil)兜底，启动新线程
    reset重置timer，重复利用timer
    stop 停止计时器的运行。stop并不会关闭channel通道，stop只能将其标记为删除状态，在适当的时候将自动删除

    Timer使用中的坑:错误创建很多的 timer，导致资源浪费
    由于Stop时不会主动关闭 C，导致程序阻塞


参考文档：
- [Golang 定时器实现原理及源码解析](https://blog.haohtml.com/archives/25962)
- [Go timer 是如何被调度的](https://cloud.tencent.com/developer/article/1840257)




