##go语言runtime包括
    Scheduler: 调度器管理所有的G,M,P，在后台执行调度循环

    Netpoll:网络轮训负责管理网络FD相关的读写、就绪事件

    Memory Management: 当代码需要内存是，负责内存分配工作

    Garbage Collector:当内存不再需要时，负责回收内存

##goruntine调度器
    G — 表示 Goroutine，它是一个待执行的任务；
    M — 表示操作系统的线程，它由操作系统的调度器调度和管理；
    P — 表示处理器，它可以被看做运行在线程上的本地调度器；
    goroutine，⼀个计算任务。由需要执⾏的代码和其上下⽂组成，上下⽂
    包括：当前代码位置，栈顶、栈底地址，状态等。
    M：machine，系统线程，执⾏实体，想要在 CPU 上执⾏代码，必须有线
    程，与 C 语⾔中的线程相同，通过系统调⽤ clone 来创建。
    P：processor，虚拟处理器，M 必须获得 P 才能执⾏代码，否则必须陷⼊休
    眠(后台监控线程除外)，你也可以将其理解为⼀种 token，有这个 token，才
    有在物理 CPU 核⼼上执⾏的权⼒。
    ***所有的执行都在M中，他是核心

    Goruntine是用户层线程
    Goruntine是go语言调度器中待执行的任务，他在运行时调度器中的地位与线程在操作系统中差不多，但是它占用了更小的内存空间，也降低了上下文切换的开销、。
    goruntine是go语言在用户态提供的线程，作为一种粒度更细的资源调度单元，sched存储了goruntine的调度相关的数据
    g 结构体关联了两个比较简单的结构体，stack 表示 goroutine 运行时的栈，光有栈还不行，至少还得包括 PC，SP 等寄存器，gobuf 就保存了这些值

    M 是操作系统线程。调度器最多可以创建 10000 个线程，但是其中大多数的线程都不会执行用户代码（可能陷入系统调用），
    最多只会有 GOMAXPROCS 个活跃线程能够正常运行
    g0 是一个运行时中比较特殊的 Goroutine，它会深度参与运行时的调度过程，包括 Goroutine 的创建、大内存分配和 CGO 函数的执行
    还存在三个与处理器相关的字段,它们分别表示正在运行代码的处理器 p、暂存的处理器 nextp 和执行系统调用之前使用线程的处理器 oldp
    还包含大量与线程状态、锁、调度、系统调用有关的字段
    
    P虚拟处理器，M 必须获得 P 才能执⾏代码，否则必须陷⼊休眠(后台监控线程除外),
    调度器在启动时就会创建 GOMAXPROCS 个处理器，所以 Go 语言程序的处理器数量一定会等于 GOMAXPROCS，这些处理器会绑定到不同的内核线程上
    其中包括与性能追踪、垃圾回收和计时器相关的字段,而 runqhead、runqtail 和 runq 三个字段表示处理器持有的运行队列，
    其中存储着待执行的 Goroutine 列表，runnext 中是线程下一个需要执行的 Goroutine。

    调度器启动，运行时通过 runtime.schedinit 初始化调度器
    创建goruntine,编译器会通过go关键字方法转换成runtime.newproc函数调用

    runtime.newproc1 会根据传入参数初始化一个 g 结构体
    1.获取或者创建新的 Goroutine 结构体；
    2.将传入的参数移到 Goroutine 的栈上；
    3.更新 Goroutine 调度相关的属性；
    总结:runtime.newproc1 会从处理器或者调度器的缓存中获取新的结构体，也可以调用 runtime.malg 函数创建

##运行队列:
    runtime.runqput 会将 Goroutine 放到运行队列上，这既可能是全局的运行队列，也可能是处理器本地的运行队列
    Go语言有两个运行队列，其中一个是处理器本地的运行队列，另一个是调度器持有的全局运行队列，只有在本地运行队列没有剩余空间时才会使用全局队列。
    每一个P结构末尾有一个runnext,当runnext有剩余空间，将goruntine设置runnext作为下一个处理器执行的任务，
    当runnext没有空间了，则在本地运行队列中寻找空间执行任务，处理器本地的运行队列是一个使用数组构成的环形链表，它最多可以存储 256 个待执行任务。
    当本地运行队列没有剩余空间，就会把本地队列中的一半和待加入的一个goruntine添加到地对应持有的全局运行队列上。

##调度循环:
    runtime.schedule进入调度循环，会从全局队列通过schedtick保证有一定几率从全局运行队列中查找对应的goruntine，
    接着从处理器本地的运行队列中查找到执行的goruntine,如果都没有还会通过findrunnable进行阻塞的查找goruntine，
    findrunnable先从本地运行队列和全局队列查找，从网络轮询器中查找等待运行的goruntine，还会通过runtime.runqsteal尝试
    从其他随机处理器中窃取待运行的goruntine，该函数可能还会窃取处理器的计时器。
    为了公平，每调用 schedule 函数 61 次就要从全局可运行 goroutine 队列中获取

    接下来由 runtime.execute 执行获取的 Goroutine，做好准备工作后，它会通过 runtime.gogo 将 Goroutine 调度到当前线程上
    多数情况下 Goroutine 在执行的过程中都会经历协作式或者抢占式调度，它会让出线程的使用权等待调度器的唤醒。
    
##触发调度
    主动挂起 — runtime.gopark -> runtime.park_m
    系统调用 — runtime.exitsyscall -> runtime.exitsyscall0
    协作式调度 — runtime.Gosched -> runtime.gosched_m -> runtime.goschedImpl
    系统监控 — runtime.sysmon -> runtime.retake -> runtime.preemptone

##主动挂起:
    runtime.gopark 是触发调度最常见的方法，该函数会将当前Goroutine暂停,被暂停的任务不会放回运行队列
    runtime.park_m 会将当前 Goroutine 的状态从 _Grunning 切换至 _Gwaiting，
    调用 runtime.dropg 移除线程和 Goroutine 之间的关联，在这之后就可以调用 runtime.schedule 触发新一轮的调度了
    runtime.ready 会将准备就绪的 Goroutine 的状态切换至 _Grunnable 并将其加入处理器的运行队列中，等待调度器的调度

##系统调用
    系统调用也会触发运行时调度器的调度，为了处理特殊的系统调用，我们甚至在 Goroutine 中加入了 _Gsyscall 状态，Go 语言通过 syscall.Syscall 和 syscall.RawSyscall 等
    使用汇编语言编写的方法封装操作系统提供的所有系统调用
    Go 语言基于协作式和信号的两种抢占式调度
##协作式调度:
    runtime.Gosched 函数会主动让出处理器，允许其他 Goroutine 运行。该函数无法挂起 Goroutine，调度器可能会将当前 Goroutine 调度到其他线程上

##系统监控
    sysmon函数不依赖P直接执行
    sysmon 执行一个无限循环，一开始每次循环休眠 20us，之后（1 ms 后）每次休眠时间倍增，最终每一轮都会休眠 10ms。
    sysmon 中会进行 netpool（获取 fd 事件）、retake（抢占）、forcegc（按时间强制执行 gc），scavenge heap（释放自由列表中多余的项减少内存占用）等处理
    总结:抢占处于系统调用的 P，让其他 m 接管它，以运行其他的 goroutine。
    将运行时间过长的 goroutine 调度出去，给其他 goroutine 运行的机会
##在四种情形下，goruntine可能会发生调度，但也并不一定发生
    1.使用了关键字go
    2.gc
    3.系统调用
    4.内存同步访问

##main goroutine 和普通 goroutine 的退出过程：
    对于 main goroutine，在执行完用户定义的 main 函数的所有代码后，直接调用 exit(0) 退出整个进程，非常霸道。
    对于普通 goroutine 则没这么“舒服”，需要经历一系列的过程。先是跳转到提前设置好的 goexit 函数的第二条指令，然后调用 runtime.goexit1，接着调用 mcall(goexit0)，
    而 mcall 函数会切换到 g0 栈，运行 goexit0 函数，清理 goroutine 的一些字段，并将其添加到 goroutine 缓存池里，然后进入 schedule 调度循环。到这里，普通 goroutine 才算完成使命。

##调度组件和调度循环
    Go 的调度流程本质上是⼀个⽣产-消费流程
    生产流程:每一个P结构末尾有一个runnext，当runnext为空，将goruntine设置runnext作为下一个处理器执行任务，
            当runnext没有空间了，在本地运行队列中寻找空间执行任务，处理器本地运行队列是一个使用数组构成的环形列表，它最多可以存储256个待执行任务
            当本地运行队列也没有空间了，就会把本地队列中的一半和待加入的一个goruntine添加到对应只有的全局队列上运行。
    消费流程:runtime.schedule进入调度循环,为了公平，每调用schedule函数61次就要从全局可运行goruntine队列中获取total/gomaxprocs+1个但不超过128个并执行，
            接着从runnext中获取goruntine，如果有值就取出来并执行，若没有则去本地运行队列查找可执行goruntine,如果有值就取出来并执行，
            若没有则去全局队列中查找，如果也没有就去网络轮询器中查找，还是没有的话会从其他处理器P中队头窃取待运行的goruntine的一半去执行，该窃取函数可能还会窃取处理器的计时器。
            获取到goruntine之后由execute执行，做好准备工作后会通过gogo将goruntine调度到当前线程上。runtime.goexit缓存相关的g结构体资源，回到schedule继续执行。
    消费整体流程:runtime.schedule->execute->gogo-exit->runtime.schedule
##处理阻塞
    channel，net read，net write，lock这些情况不会阻塞调度循环，而是把goruntine挂起，就是让g先进某个数据结构待ready后再进行执行，不会占用线程。
    这时候，线程会进入schedule，继续消费队列，执行其他的g
    引用阻塞在lock上，g按照lock addr排列的二叉搜索树，按ticket排列的小顶堆，ticket每个sudog初始化时用fastrand生成，
    树上的每一个节点都是一个链表。
    为啥有的等待是 sudog，有的是 g?
    ⼀个 g 可能对应多个 sudog，⽐如⼀个 g 会同时 select 多个channel


    gobuf 描述⼀个 goroutine 所有现 场，从⼀个 g 切换到另⼀个 g， 只要把这⼏个现场字段保存下来，
    再把 g 往队列⾥⼀扔， m 就可以执⾏其它 g 了

参考文档：
- [Golang 调度器 GMP 原理与调度全分析](https://learnku.com/articles/41728)
- [图解Go运行时调度器](https://tonybai.com/2020/03/21/illustrated-tales-of-go-runtime-scheduler/)
- [面向信仰编程](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-goroutine/)
- [码农桃花源goroutine 调度器](https://qcrao91.gitbook.io/go/goroutine-tiao-du-qi)
- [Goroutine调度策略](https://mp.weixin.qq.com/s/2objs5JrlnKnwFbF4a2z2g)


