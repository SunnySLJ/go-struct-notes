##go编译原理

    go run -x main.go 查看编译过程
    编译 go tool compile -S  ./hello.go | grep "hello.go:5"
    反编译 https://golang.org/ref/spec
    go  tool objdump ./hello.o反编译寻找make实现


##调试工具
     dlv exec ./hello
     调试时可以用b 、 s、si、n、c（continue从一个断点到下一个断点）、disass（反汇编）
     bp


##逃逸分析
    go build -gcflags="-m" 查看为什么会逃逸。
    逃逸分析实现原理，大致算法基于两个不变性：
    a.指向栈对象的指针不能存储在堆中
    b.指向栈对象的指针不能超过该栈对象的存活期。


##内存管理
    内存通过自动allocator 手工分配
    内存需要自动collector 手工回收
    go的话就是有allocator和collector自动分配与自动回收，这种自动回收的就是垃圾回收技术

##内存管理器三个角色
    Mutator：对象图。fancy(花哨的)word for application,其实就是你写的引用程序它不断的修改对象的引用关系。
    Allocator：内存分配器，负责管理从操作系统中分配出的内存空间，malloc其实底层就有一个内存分配器的实现(glibc中)，tcmalloc是malloc多线程改进版。
              Go中的实现类似tcmalloc。tcmalloc是需要加锁的。
    Collector：垃圾收集器，负责清理死对象，释放内存空间
    Mutator用户程序申请内存时，他会通过Allocator内存分配器申请新的内存，而分配器会负责从堆中初始化相应的内存区域。
