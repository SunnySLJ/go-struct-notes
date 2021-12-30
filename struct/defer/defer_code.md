##defer分析以及总结
    runtime._defer 结构体是延迟调用链表上的一个元素，所有的结构体都会通过 link 字段串联成链表
    
    Golang官方博客里总结了defer的三条行为规则
    规则一：延迟函数的参数在defer语句出现时就已经确定下来了
    规则二：延迟函数执行按后进先出顺序执行，即先出现的defer最后执行
    规则三：延迟函数可能操作主函数的具名返回值

    return不是原子操作，执行过程是: 保存返回值(若有)—>执行defer（若有）—>执行ret跳转
    defer关键字首先会调用runtime.deferproc 定义一个延迟调用对象，然后再函数结束前，调用runtime.deferreturn来完成defer定义的函数的调用

    deferproc创建一个延迟执行的函数，并将这个延迟函数挂在当前g的_defer的链表上
    newdefer的作用是获取一个*_defer*对象， 并推入 g._defer链表的头部

## 总结:

--[深入理解Go-defer的原理剖析](https://juejin.cn/post/6844903936508297223)




