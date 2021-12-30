##netpoll分析以及总结
    netpoller 网络轮询器

##I/O模型
    操作系统中包含阻塞 I/O、非阻塞 I/O、信号驱动 I/O 与异步 I/O 以及 I/O 多路复用五种 I/O 模型

    阻塞I/O:
    操作系统中多数的I/O操作都是阻塞请求，一旦执行 I/O 操作，应用程序会陷入阻塞等待I/O操作的结束。

    非阻塞I/O:
    当进程把一个文件描述符设置成非阻塞时，执行 read 和 write 等 I/O 操作会立刻返回
    第一次从文件描述符中读取数据会触发系统调用并返回 EAGAIN 错误，EAGAIN 意味着该文件描述符还在等待缓冲区中的数据；
    随后，应用程序会不断轮询调用 read 直到它的返回值大于 0，这时应用程序就可以对读取操作系统缓冲区中的数据并进行操作

    I/O多路复用:
    I/O 多路复用被用来处理同一个事件循环中的多个 I/O 事件
    多路复用函数会阻塞的监听一组文件描述符，当文件描述符的状态转变为可读或者可写时
    select 会返回可读或者可写事件的个数，应用程序可以在输入的文件描述符中查找哪些可读或者可写，然后执行相应的操作。
    I/O 多路复用模型是效率较高的 I/O 模型，它可以同时阻塞监听了一组文件描述符的状态
    很多高性能的服务和应用程序都会使用这一模型来处理 I/O 操作，例如：Redis 和 Nginx 等。

##多模块
    Go语言在网络轮询器中使用I/O多路复用模型处理I/O操作，但是他没有选择最常见的系统调用select
    Go 语言为了提高在不同操作系统上的 I/O 操作性能，使用平台特定的函数实现了多个版本的网络轮询模块
    如果目标平台是 Linux，那么就会根据文件中的 // +build linux 编译指令选择 src/runtime/netpoll_epoll.go 
    并使用epoll函数处理用户的 I/O 操作

##接口
    runtime.netpollinit — 初始化网络轮询器，通过 sync.Once 和 netpollInited 变量保证函数只会调用一次；
    runtime.netpollopen — 监听文件描述符上的边缘触发事件，创建事件并加入监听；
    runtime.netpoll — 轮询网络并返回一组已经准备就绪的 Goroutine，传入的参数会决定它的行为3；
                      如果参数小于 0，无限期等待文件描述符就绪；
                      如果参数等于 0，非阻塞地轮询网络；
                      如果参数大于 0，阻塞特定时间轮询网络；
    runtime.netpollBreak — 唤醒网络轮询器，例如：计时器向前修改时间时会通过该函数中断网络轮询器4；
    runtime.netpollIsPollDescriptor — 判断文件描述符是否被轮询器使用；

##数据结构
    pollDesc结构体，它会封装操作系统的文件描述符
    rseq 和 wseq — 表示文件描述符被重用或者计时器被重置5；
    rg 和 wg — 表示二进制的信号量，可能为 pdReady、pdWait、等待文件描述符可读或者可写的 Goroutine 以及 nil；
    rd 和 wd — 等待文件描述符可读或者可写的截止日期；
    rt 和 wt — 用于等待文件描述符的计时器
    runtime.pollDesc 结构体会使用 link 字段串联成链表存储在 runtime.pollCache 中

    netFD 是 Golang 自己封装的一个“网络文件描述符”结构
    用户可以通过epoll_ctl调用将fd注册到epoll实例上，而epoll_wait则会阻塞监听所有的epoll实例上所有的fd的事件
    通过epoll_ctl注册fd，一个fd只完成一次从用户态到内核态的拷贝而不需要每次调用时都拷贝一次，并且epoll使用红黑树存储所有的fd因此重复注册是没用的


##多路复用
    netpoller实际上是对I/O多路复用技术的封装:
##netpoll的初始化
    runtime.netpollGenericInit 会调用平台上特定实现的 runtime.netpollinit
    调用epollcreate1创建一个新的epoll文件描述符，这个文件描述符会在整个程序的生命周期中使用
    通过runtime.nonblockingPipe创建一个用于通信的管道
    使用epollctl将用于读取数据的文件描述符打包成epollevent事件加入监听
    runtime.netpollBreak 会向管道中写入数据唤醒 epoll
    因为目前的计时器由网络轮询器管理和触发，它能够让netpoll立刻返回并让运行时检查是否有需要触发的计时器
##轮训事件    
    pollDesc.init初始化还会通过runtime.poll_runtime_pollOpen重置轮训信息runtime.pollDesc
    并调用netpollopen初始化轮训事件，netpollopen调用epollctl向全局的轮训文件描述符epfd中加入新的
    轮训事件监听文件描述的可读和可写状态 
##事件循环
    当我们在文件描述符上执行读写操作时，如果文件描述符不可读或者不可写，当前 Goroutine 会执行 
    runtime.poll_runtime_pollWait 检查 runtime.pollDesc 的状态并调用 runtime.netpollblock 等待文件描述符的可读或者可写
    runtime.netpollblock 是 Goroutine 等待 I/O 事件的关键函数，它会使用运行时提供的 runtime.gopark 让出当前线程，
    将 Goroutine 转换到休眠状态并等待运行时的唤醒
    
    轮训等待，go运行时会在调度或者系统监控中调用runtime.netpoll轮训网络，
    根据传入的delay计算epoll系统调用需要等待的时间。
    调用epollwait等待可读或者可写事件的发生
    当epollwait系统调用返回的值大于0时意味着被监控的文件描述符出现了待处理的时间，
    在循环中依次处理epollevent事件，处理的事件总共包含两种
    一种是调用runtime.netpollbreak触发的事件，该函数的作用是终端netpoll
    另一种是其他文件描述符的正常读写事件交给netpollready处理
    将 runtime.pollDesc 中的读或者写信号量转换成 pdReady 并返回其中存储的 Goroutine;
    如果返回的 Goroutine 不会为空，那么运行时会将该 Goroutine 会加入 toRun 列表，并将列表中的全部 Goroutine 加入运行队列并等待调度器的调度。

    3.如何从网络轮询器获取触发的事件

##截止日期
    网络轮询器和计时器关系非常紧密，不仅网络轮询器负责计时器的唤醒，还因为文件和网络 I/O 的截止日期也由网络轮询器负责处理。
    截止日期在 I/O 操作中，尤其是网络调用中很关键，网络请求存在很高的不确定因素，我们需要设置一个截止日期保证程序的正常运行
    
##原理
    1.当调用epoll_create，其实是创建了一个eventpoll结构体对象，在epoll运行期间的相关数据都存在此结构里面。
    2.接着是通过epoll_ctl注册socket s 感兴趣的事件，结构中的rbr就是用来存放所有注册的socket。同时epoll_ctl接口还会注册回调函数ep_poll_callback。
    3.网卡收到数据后，会把数据复制到内核空间，并触发回调函数ep_poll_callback，ep_poll_callback会把就绪的fd指针放入rdllist，
      并检查wq中是否有阻塞的线程，如果有则唤醒它们。
    4.调用epoll_wait函数检查是否有事件触发(就绪)，如果有，则通过参数2返回（这里其实就是检查rdllist是否为空，如果不为空则返回事件列表）。
      参数4为阻塞时间，若不为0，在rdllist为空时，调用epoll_wait的线程会被阻塞，并放到wq中，如果阻塞时间结束，仍然没有事件发生，则被唤醒；
      如果等待期间有事件发生内核触发ep_poll_callback回调并唤醒这个fd上阻塞的线程。

##linux中epoll的I/O多路复用
    epoll 是 Linux kernel 2.6 之后引入的新 I/O 事件驱动技术
    epoll 的 API 非常简洁，涉及到的只有 3 个系统调用
    epoll_create、epoll_ctl、epoll_wait
    epoll_create 创建一个 epoll 实例并返回 epollfd
    epoll_ctl 注册 file descriptor 等待的 I/O 事件(比如 EPOLLIN、EPOLLOUT 等) 到 epoll 实例上
    epoll_wait 则是阻塞监听 epoll 实例上所有的 file descriptor 的 I/O 事件，它接收一个用户空间上的一块内存地址 (events 数组)
               kernel 会在有 I/O 事件发生的时候把文件描述符列表复制到这块内存地址上，然后 epoll_wait 解除阻塞并返回，最后用户空间上的程序就可以对相应的 fd 进行读写了
    epoll 采用红黑树来存储所有监听的 fd，而红黑树本身插入和删除性能比较稳定，时间复杂度 O(logN)。通过 epoll_ctl 函数添加进来的 fd 都会被放在红黑树的某个节点内，所以，重复添加是没有用的

##go netpoller基本原理
    通过在底层对 epoll/kqueue/iocp 的封装，从而实现了使用同步编程模式达到异步执行的效果。
    总结来说，所有的网络操作都以网络描述符 netFD 为中心实现。netFD 与底层 PollDesc 结构绑定，当在一个 netFD 上读写遇到 EAGAIN 错误时，
    就将当前 goroutine 存储到这个 netFD 对应的 PollDesc 中，同时调用 gopark 把当前 goroutine 给 park 住，直到这个 netFD 上再次发生读写事件，
    才将此 goroutine 给 ready 激活重新运行。显然，在底层通知 goroutine 再次发生读写等事件的方式就是 epoll/kqueue/iocp 等事件驱动机制

    netpoll 是通过 park goroutine 从而达到阻塞 Accept/Read/Write 的效果，而通过调用 gopark，
    goroutine 会被放置在某个等待队列中，这里是放到了 epoll 的 "interest list" 里，底层数据结构是由红黑树实现的 eventpoll.rbr，
    此时 G 的状态由 _Grunning为_Gwaitting ，因此 G 必须被手动唤醒(通过 goready )，否则会丢失任务，应用层阻塞通常使用这种方式

    client 连接 server 的时候，listener 通过 accept 调用接收新 connection，每一个新 connection 都启动一个 goroutine 处理，accept 调用会把该 connection 的 fd 连带所在的 goroutine 上下文信息封装注册到 epoll 的监听列表里去，当 goroutine 调用 conn.Read 或者 conn.Write 等需要阻塞等待的函数时，会被 gopark 给封存起来并使之休眠，让 P 去执行本地调度队列里的下一个可执行的 goroutine，往后 Go scheduler 会在循环调度的 runtime.schedule() 函数以及 sysmon 监控线程中调用 runtime.netpoll 以获取可运行的 goroutine 列表并通过调用 injectglist 把剩下的 g 放入全局调度队列或者当前 P 本地调度队列去重新执行。

    那么当 I/O 事件发生之后，netpoller 是通过什么方式唤醒那些在 I/O wait 的 goroutine 的？答案是通过 runtime.netpoll。

    runtime.netpoll 的核心逻辑是
    1.根据调用方的入参 delay，设置对应的调用 epollwait 的 timeout 值；
    2.调用 epollwait 等待发生了可读/可写事件的 fd；
    3.循环 epollwait 返回的事件列表，处理对应的事件类型， 组装可运行的 goroutine 链表并返回。

    Go 在多种场景下都可能会调用 netpoll 检查文件描述符状态，netpoll 里会调用 epoll_wait 从 epoll 的 eventpoll.rdllist 
    就绪双向链表返回，从而得到 I/O 就绪的 socket fd 列表，并根据取出最初调用 epoll_ctl 时保存的上下文信息，恢复 g。所以执行完netpoll 之后，会返回一个就绪 fd 列表对应的 goroutine 链表，接下来将就绪的 goroutine 通过调用 injectglist 加入到全局调度队列或者 P 的本地调度队列中，启动 M 绑定 P 去执行。

    具体调用 netpoll 的地方，首先在 Go runtime scheduler 循环调度 goroutines 之时就有可能会调用 netpoll 
    获取到已就绪的 fd 对应的 goroutine 来调度执行。

    首先 Go scheduler 的核心方法 runtime.schedule() 里会调用一个叫 runtime.findrunable() 的方法
    获取可运行的 goroutine 来执行，而在 runtime.findrunable() 方法里就调用了 runtime.netpoll 
    获取已就绪的 fd 列表对应的 goroutine 列表：

    sysmon 监控线程会在循环过程中检查距离上一次 runtime.netpoll 被调用是否超过了 10ms，
    若是则会去调用它拿到可运行的 goroutine 列表并通过调用 injectglist 把 g 列表放入全局调度队列或者当前 P 本地调度队列等待被执行：
    综上，Go 借助于 epoll/kqueue/iocp 和 runtime scheduler 等的帮助，设计出了自己的 I/O 多路复用 netpoller，成功地让 Listener.Accept / conn.Read / conn.Write 等方法从开发者的角度看来是同步模式

    Go netpoller 的问题
    Go netpoller 的设计不可谓不精巧、性能也不可谓不高，配合 goroutine 开发网络应用的时候就一个字：爽。因此 Go 的网络编程模式是及其简洁高效的，
    然而，没有任何一种设计和架构是完美的， goroutine-per-connection 这种模式虽然简单高效，但是在某些极端的场景下也会暴露出问题：goroutine 虽然非常轻量，
    它的自定义栈内存初始值仅为 2KB，后面按需扩容；海量连接的业务场景下， goroutine-per-connection ，此时 goroutine 数量以及消耗的资源就会呈线性趋势暴涨，
    虽然 Go scheduler 内部做了 g 的缓存链表，可以一定程度上缓解高频创建销毁 goroutine 的压力，但是对于瞬时性暴涨的长连接场景就无能为力了，
    大量的 goroutines 会被不断创建出来，从而对 Go runtime scheduler 造成极大的调度压力和侵占系统资源，然后资源被侵占又反过来影响 Go scheduler 的调度，进而导致性能下降。

    网络轮询器实际上是对 I/O 多路复用技术的封装
    运行时的调度器和系统调用都会通过 runtime.netpoll 与网络轮询器交换消息，获取待执行的 Goroutine 列表，
    并将待执行的 Goroutine 加入运行队列等待处理。所有的文件 I/O、网络 I/O 和计时器都是由网络轮询器管理的，它是 Go 语言运行时重要的组件。

##总结:
    epollcreate创建一个epoll实例并返回epollfd
    epollcl注册fd到epoll实例上
    epollwait阻塞监听epoll上所有I/O事件，接受一个用户空间上的一块内存地址，有事件发生是把文件描述符列表复制到这块内存地址上，epollwait解除阻塞并返回。
    
    1.addTimer或者每一个pollDesc初始化都会调用netpollGenericInit初始化，
      创建一个新的epoll文件描述符，创建一个通信的管道，通过epollctl调用将fd注册到epoll实例并将epoll打包成epollevent事件加入监听，
      epollctl注册fd，一个fd只完成一次从用户态到内核态的拷贝，epoll使用红黑树存储所有fd。
      pollDesc初始化还会重置轮休信息并调用初始化轮休监听文件描述的可读和可写状态事件，
    2.当前goruntine执行pollWait检查pollDesc状态等待文件描述符的可读或者可写，若不可读或不可写，会通过gopark出让当前线程等待被唤醒
    3.go运行时会在调度或系统监控中调用netpoll
      事件通过netpoll唤醒I/O wait的goruntine
      根据传入的delay计算epoll系统调用需要等待的时间，调用epollwait等待，
      epollwait是阻塞监听，event中的data就是pollDesc，根据ready事件是read或write分别从pollDesc的rg、wg上拿到该唤醒的goruntine，
      已经ready的goruntine push到toRun链表，最终调用injectglist将toRun链表中全部goruntine加入到全局队列等待调度器的调度
    4.pollDesc会使用link字段串联成链表存储在pollCache中，初始化链表会调用persistentalloc为这些数据结构分配不会被GC回收的内存，
      确保被epoll和kqueue在内核空间去引用
    5.sysmon监控线程会在循环过程中检查距离上一次netpoll被调用是否超过了10ms，若是则回去调用它
      拿到可运行的goruntine列表并通过indectglist把g列表放入全局调度队列或者当前P本地队列等待被执行
    6.缺点:对于瞬间暴涨的长连接场景中大量的goruntine被创建销毁对scheduler造成的调度压力和侵占系统资源是很不利的。进而导致性能下降。
参考文档：
- [网络轮询器](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-netpoller/)
- [Go netpoller 原生网络模型之源码全面揭秘](https://strikefreedom.top/go-netpoll-io-multiplexing-reactor)
- [Golang Netpoll 实现原理](https://lifan.tech/2020/07/03/golang/netpoll/)



