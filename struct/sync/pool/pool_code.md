##pool源码分析以及总结
    Pool由noCopy、local、victim、New四个变量组成
    noCopy用于防止Pool被值复制使用，可使用go vet工具进行Pool值复制的检查；
    local是[]poolLocal类型的指针，用于保存临时对象；[]poolLocal的长度等于处理器P的长度，且每一个处理器P对应一个poolLocal；
    victim也是[]poolLocal类型的指针；Pool定期会调用poolCleanUp()进行清理操作，victim会保存上一轮的local；
    New指定了新对象的创建方法；从空的Pool中获取对象，会调用New方法创建新对象。

    poollocal中的shared是一个poolChain并发安全的双向链表，用于保存临时对象，可被所有处理器P访问
    Get()方法用于从Pool中获取空闲对象,本地local的private->本地local的shared->其他处理器的local的shared
    ->victim->New
    Put()用于存放空闲对象，其存放顺序如下：本地local的private->本地local的shared。

    程序内的所有Pool对象都会保存在allPools和oldPools这两个全局变量中。
    其中，allPools用于保存local非空的Pool对象，oldPools用于保存local为空、victim非空的Pool对象。
    poolCleanup()会对所有Pool对象进行清理操作，poolCleanup()的调用时刻在GC开始时。
    poolCleanup()执行时，GC会进行STW操作，所以poolCleanup()无需并发控制。

    在执行Get()方法和Put()方法时，都会执行pin()方法。pin()方法主要进行了两个操作：
    通过 runtime_procPin()将协程对应的线程设置为不可抢占状态，防止协程被抢占；
    当Pool的local需要被创建时，进行local的创建。

##分析
    两个场景使用，进程中inuse_object数过多，gc mark消耗大量CPU
    进程中inuse_object数过多，进程RSS占用过高
    可以优化在生命周期开始前启用sync.Pool.get(),请求结束是pool.put();
    **因为var pool = sync.Pool{}的时候会生成Pool并append到全局runtime.allPool，写应用层的时候也不能一直去申请pool，
    这也会导致runtime.allPools去append和加锁。也可能导致非常严重的性能问题。
    sync.Pool为每个P分配一个本地池，当执行Get或者Put操作的时候，会先将goroutine和某个P的子池关联，再对该子池进行操作。
    每个P的子池分为私有对象和共享列表对象，私有对象只能被指定P访问，共享队列shared可以被任何P访问

##总结 当进程中已分配但未释放的对象数过多或者gc mark消耗大量CPU的时候使用sync.Pool，当然也因为生成sync.Pool对象的时候会append到全局allPool中，
##    写应用层的时候也不能一直申请pool，这样会导致一直去allPools中append和加锁。
##    sync.Pool为每个P分配一个本地池，当执行Get或者Put操作的时候，会先将goroutine和某个P的子池关联，再对该子池进行操作。
##    每个P的子池分为私有对象和共享列表对象，私有对象只能被指定P访问，共享队列shared可以被任何P访问

1. sync.Pool的核心作用 - 读源码，`缓存稍后会频繁使用的对象`+`减轻GC压力`
2. sync.Pool的Put与Get - Put的顺序为`local private-> local shared`，Get的顺序为 `local private -> local shared -> remote shared`
3. 思考sync.Pool应用的核心场景 - `高频使用且生命周期短的对象，且初始化始终一致`，如fmt
4. 探索Go1.13引入`victim`的作用 - 了解`victim cache`的机制

参考文档：
- [golang Pool源码解析）](https://juejin.cn/post/6966018400430587935)

 


