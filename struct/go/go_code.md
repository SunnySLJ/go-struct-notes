##  go run -x main.go 查看编译过程

##编译过程
    对go源文件进行词法分析、语法分析、语义分析、中间代码生成、中间代码优化，最后生成汇编代码文件，以.s作为文件后缀。
    语义分析将语法变成2抽象语法树AST，并在树上做类型检查。
    中间代码表示形式为SSA(Static Single-Assignment，静态单赋值),之所以称之为单赋值，是因为每个名字在SSA中仅被赋值一次
    中间代码生成过程就是从AST抽象语法树到SSA中间代码的过程。最后生成能在不同CPU架构上运行的代码。

##编译与反编译工具
    编译 go tool compile -S  ./hello.go | grep "hello.go:5"
    go build hello.go && go tool objdump ./hello
    反编译 go tool objdump ./hello.o

    go build -gcflags -S main.go
    其他方法go build -gcflags "-N -l" -o hello
    进入gdb调试模式，执行info files得到可执行文件
    通过 readelf -H 中的 entry 找到程序⼊⼝

##使用调试工具dlv
    dlv debug hello.go
    dlv exec ./hello
    调试时可以用b、s、si、n、c（continue从一个断点到下一个断点）、disass（反汇编）、bp

    break（b）main.main #在main包里的main函数入口打断点
    continue（c） #继续运行，直到断点处停止
    next（n） #单步运行
    locals #打印local variables
    print（p） #打印一个变量或者表达式
    restart（r） #Restart Process

##go语言的runtime整个Go语言都围绕这以上四个板块
    包括Scheduler、Netpoll、Memory Management、Garbage Collector
    Scheduler: 调度器管理所有的G,M,P，在后台执行调度循环
    Netpoll:网络轮训负责管理网络FD相关的读写、就绪事件
    Memory Management: 当代码需要内存是，负责内存分配工作
    Garbage Collector:当内存不再需要时，负责回收内存


##go调度过程本质上是一个生产，一个消费
    runtime中有一个本地队列和全局队列，还有runnext。
    调度流程是所有调度循环的核心逻辑。
    每一个P结构末尾有一个runnext
    1.goruntine的生产端
    2.goruntine的消费端
    分为四部分schedule、runtime.execute、runtime.gogo、runtime.goexit
    schedule每61次从全局队列中获取goruntine。

    3.看文字定义
    G：goroutine，⼀个计算任务。由需要执⾏的代码和其上下⽂组成，上下⽂
    包括：当前代码位置，栈顶、栈底地址，状态等。
    M：machine，系统线程，执⾏实体，想要在 CPU 上执⾏代码，必须有线
    程，与 C 语⾔中的线程相同，通过系统调⽤ clone 来创建。
    P：processor，虚拟处理器，M 必须获得 P 才能执⾏代码，否则必须陷⼊休眠(后台监控线程除外)，
    你也可以将其理解为⼀种 token，有这个 token，才 有在物理 CPU 核⼼上执⾏的权⼒。

## 处理阻塞
    为什么阻塞的等待有的是sudog，有的是g
    ⼀个 g 可能对应多个 sudog，⽐如⼀个 g 会同时 select 多个channel

参考文档：
- [Go程序是怎样跑起来的](https://www.cnblogs.com/sunsky303/p/11131263.html)
- [如何使用Go的调试工具dlv](https://zhuanlan.zhihu.com/p/126183467)



