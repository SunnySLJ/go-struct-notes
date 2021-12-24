##channel源码分析以及总结
    channel底层数据结构hchan，数据结构固定长度的双向循环列表

    qcount   uint       // 当前队列中总元素个数
    dataqsiz uint       // 环形队列长度，即缓冲区大小（申明channel时指定的大小）     
    buf unsafe.Pointer  // 环形队列指针
    elemsize            // buf中每个元素的大小
    closed              // 当前通道是否处于关闭状态，创建通道时该字段为0，关闭时字段为1
    elemtype *_type     // 元素类型，用于传值过程的赋值
    sendx    uint       // 环形缓冲区中已发送位置索引
    recvx    uint       // 环形缓冲区中已接收位置索引
    recvq    waitq      // 等待读消息的groutine队列
    sendq    waitq      // 等待写消息的groutine队列
    lock mutex          // 互斥锁，为每个读写操作锁定通道（发送和接收必须互斥）


    1.首先make一个有缓存的channel，channel中包含buf、sendx、recvx、sendq、recvq、lock
    2.channel接收传入的值，当buf有空，添加到buf。当buf满了，继续传值，会通过getg方法获取一个g和hchan生成一个sudog，放到sendq里面。继续接收就继续生成sudog放到sendq里。等待被读goruntine唤醒。
    3.接收的时候recvq消费，当buf中有值，先进先出，先消费，然后将sendq中的sudog放到buf的头中，继续消费。
    4.总结:向channel写数据
       开始发送，recvq为空，如果buf有空位，将数据写入buf队尾，如果没有空位，加入到sendq生成sudog等待被唤醒，被唤醒是被取走，如果recvq不为空，从recvq取出一个g，
       将数据写入g，唤醒g从channel读数据
      开始接收，sendq非空，有缓冲区buf的话，先从buf队首取一个元素，从sendq取一个g放到buf队尾，没有缓冲区，从sendq取出g读取数据，接下来唤醒g，
    如果sendq为空，qcount>0即buf中有值，读出buf中元素，buf中也没有，则将当前goruntine加入recvq等待被唤醒，被唤醒时数据写入。
    5.关闭channel时会把recvq中g全部唤醒，本该写入g的数据位置为nil，把sendq中G全部唤醒，但是这个g会panic。关闭值为nil的channel、关闭已经关闭channel、向已经关闭的channel写数据这些都会panic
    6.可接管的阻塞，均有gopark挂起，每个gopark对应一个唤醒
     channel send->channel recv/close
     lock -> unLock
     read-> read ready,epoll_wait返回该fd事件
    timer-> checkTimers,检查到期唤醒

    如果recvq不为空，从recvq中取出一个等待接收数据的Groutine，将数据发送给该Groutine
    如果recvq为空，才将数据放入buf中
    如果buf已满，则将要发送的数据和当前的Groutine打包成Sudog对象放入sendq，并将groutine置为等待状态

    如果有等待发送数据的groutine，从sendq中取出一个等待发送数据的Groutine，取出数据
    如果没有等待的groutine，且环形队列中有数据，从队列中取出数据
    如果没有等待的groutine，且环形队列中也没有数据，则阻塞该Groutine，并将groutine打包为sudogo加入到recevq等待队列中

    gopark函数做的主要事情分为两点：
    解除当前goroutine的m的绑定关系，将当前goroutine状态机切换为等待状态；
    调用一次schedule()函数，在局部调度器P发起一轮新的调度。


## 总结:
    非缓冲channel如果没有goroutine读取接受者，那么发送者会一直阻塞，缓冲channel类似一个队列，只有队列满之后才可能阻塞。
    sendchan方法，recvq为空，如果buf有空位，则将数据写入buf对位，如果buf满了，则将加入到sendq生成sudog等待被唤醒，如果recvq不为空，从recvq去一个g，
    将数据写入g并唤醒g从channel去读数据
    chanrecv方法，sendq非空，有缓存区buf的话，先从buf队首取一个元素，从sendq取一个g放到buf队尾，没有buf，从sendq取出g读取数据唤醒g，
    如果sendq为空，qcount>0即buf中有值，读出buf中的元素，若buf中也没有，则将当前goruntine加入recvq等待被唤醒，被唤醒时数据被写入
    closechan方法，上锁，把recvq中g全部唤醒，本该写入g的数据位置设为nil，把sendq中G全部唤醒。
    关闭值为nil的channel，关闭已经关闭的channel，向已经关闭的chaannel写数据这些都会panic
    

1. channel用于Goroutine间通信时的注意点 - 合理设置channel的size大小 / 正确地`关闭channel`
2. 合理地运用channel的发送与接收 - 运用函数传入参数的定义，限制 `<- chan` 和 `chan <-`
3. channel的底层实现 - `环形队列`+`发送、接收的waiter通知`，结合Goroutine的调度思考
4. 理解并运用channel的阻塞逻辑 - 理解channel的每一对 `收与发` 之间的逻辑，巧妙地使用
5. 思考channel嵌套后的实现逻辑 - 理解用 `chan chan` 是怎么实现 `两层通知` 的？


参考文档：
- [golang channel从源码分析实现原理（go 1.14）](https://juejin.cn/post/6875325172249788429)
- [Golang channel 源码深度剖析](https://www.cyhone.com/articles/analysis-of-golang-channel/)
- [gopark函数和goready函数原理分析](https://blog.csdn.net/u010853261/article/details/85887948)



