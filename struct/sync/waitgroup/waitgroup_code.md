##waitgroup源码分析以及总结
    waitgroup结构体保函nocopy，state1。
    state1高32位是计数器，低32位是waiter计数，当state没有按照8对齐时进行偏4个字节来使用
    add方法给计数器增加delta，delta可能为负值
    done就是将计数器减1，如果计数器为0则触发runtime_Semrelease唤醒所有阻塞在Wait上的g
    wait会阻塞知道wg的counter变为0，底层就是使用CAS，如果counter没到0就调用runtime_Semacquire挂起，
    其中都是通过信号量sema挂起和唤醒

##注意
    waitgroup不要进行copy，add要在goroutine前执行
    逻辑复杂时建议用defer保证调用
    适用`并发的Goroutine执行的逻辑相同时`的场景,否则代码并不整洁

