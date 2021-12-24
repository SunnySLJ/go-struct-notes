##context源码分析以及总结

    用一句话解释 Context 在 Go 语言中的作用就是:
    Context 为同一任务的多个 goroutine 之间提供了 退出信号通知 和 元数据传递的功能。

    主要用来在goroutine之间传递上下文信息，包括：取消信号、超时时间、截止时间、k-v 等
    context的适用机制,上层任务取消后，所有的下层任务都会被取消。
    中间某一层的任务取消后，只会将当前任务的下层任务取消，而不会影响上层的任务以及同级任务

    在未考虑清楚是否传递、如何传递context时用TODO，作为发起点的时候用Background
    context是并发安全的
    context可以进行传值，但是在使用context进行传值的时候我们应该慎用，使用context传值是一个比较差的设计，
    比较常见的使用场景是传递请求对应用户的认证令牌以及用于进行分布式追踪的请求 ID。
    对于context的传值查询，context查找的时候是向上查找，找到离得最近的一个父节点里面挂载的值，所以context在查找的时候会存在覆盖的情况，
    如果一个处理过程中，有若干个函数和若干个子协程。在不同的地方向里面塞值进去，对于取值可能取到的不是自己放进去的值。
    当使用 context 作为函数参数时，直接把它放在第一个参数的位置，并且命名为 ctx。另外，不要把 context 嵌套在自定义的类型里。

1. Context上下文 - 结合Linux操作系统的`CPU上下文切换/子进程与父进程`进行理解
2. 如何优雅地使用context - 与`select`配合使用，管理协程的生命周期
3. Context的底层实现是什么？ - `mutex`与`channel`的结合，前者用于初始部分参数，后者用于通信

参考文档：
- [go中context源码刨铣](https://boilingfrog.github.io/2021/02/22/go%E4%B8%ADcontext/)
- [Context是怎么在Go语言中发挥关键作用的](https://juejin.cn/post/7000300756116963342)
- [深度解密Go语言之context](https://mp.weixin.qq.com/s/GpVy1eB5Cz_t-dhVC6BJNw)

  

