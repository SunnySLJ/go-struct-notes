##panic分析以及总结
    1.编译器会负责做转换关键字的工作；
      a.将 panic 和 recover 分别转换成 runtime.gopanic 和 runtime.gorecover；
      b.将 defer 转换成 runtime.deferproc 函数；
      c.在调用 defer 的函数末尾调用 runtime.deferreturn 函数；
    2.在运行过程中遇到 runtime.gopanic 方法时，会从 Goroutine 的链表依次取出 runtime._defer 结构体并执行；
    3.如果调用延迟执行函数时遇到了 runtime.gorecover 就会将 _panic.recovered 标记成 true 并返回 panic 的参数；
      a.在这次调用结束之后，runtime.gopanic 会从 runtime._defer 结构体中取出程序计数器 pc 和栈指针 sp 并调用 runtime.recovery 函数进行恢复程序；
      b.runtime.recovery 会根据传入的 pc 和 sp 跳转回 runtime.deferproc；
      c.编译器自动生成的代码会发现 runtime.deferproc 的返回值不为 0，这时会跳回 runtime.deferreturn 并恢复到正常的执行流程；
    4.如果没有遇到 runtime.gorecover 就会依次遍历所有的 runtime._defer，并在最后调用 runtime.fatalpanic 中止程序、
      打印 panic 的参数并返回错误码 2；






