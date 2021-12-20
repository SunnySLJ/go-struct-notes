##atomic源码分析以及总结

##详解
    go中atomic类操作最终是使用assembly进行cpu指令的调用实现的。
    doc.go内只是对函数的定义，具体实现在asm文件中，asm函数实现通过JMP跳转指令捅一刀runtime下的各个函数进行
    通过cmd/compile/internal/gc/ssa.go的函数alias得知go中函数对应的指定函数别名
    所以atomic最终实现就是汇编指令操作cpu，通过lock cmpxchg指令

    lock 指令前缀可以使许多指令操作（ADD, ADC, AND, BTC, BTR, BTS, CMPXCHG, CMPXCH8B, DEC, INC, NEG, NOT, OR, SBB, SUB, XOR, XADD, and XCHG）变成原子操作。CMPXCHG 指令用来实现 CAS 操作。
    atomic.CompareAndSwap 即是使用 lock cmpxchg 来实现的。

##atomic.Store
    atomic.Store源码，通过Store保存后类型就固定下来了，后续操作必须使用相同类型，否则会panic，且不能保存nil。
    首次使用会调用runtime_procPin()禁止当前P被强占，然后调用CAS强占乐观锁。如果抢到了就去修改typ和data
