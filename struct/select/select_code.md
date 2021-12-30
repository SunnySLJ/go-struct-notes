##select分析以及总结
    hselect结构体select 的本体数据结构
    scase结构体，包含hchan存储case中使用的channel，还有一个elem

    go语言的select与操作系统中的select比较相似。
    go中select能让goruntine同时等待多个channel刻度或者可写，在多个文件或者channel状态改变之前，
        select会一直阻塞当前线程或者goruntine

    非阻塞的收发，通常情况下，select会阻塞当前goruntine并等待多个channel中一个达到可以收发的状态。
    如果select控制结构中包含default，会有两种情况，当存在可以收发的channel，直接处理该channel
    对应的case，当不存在可以收发的channel，会直接执行default

    首先在编译期间，Go 语言会对 select 语句进行优化，它会根据 select 中 case 的不同选择不同的优化路径
    空的 select 语句会被转换成调用 runtime.block 直接挂起当前 Goroutine
    如果select语句只包含一个case，首先判断操作的channel是否为空，然后执行case结构中的内容
    如果select语句中只包含两个case并其中一个是default，会非阻塞地执行收发操作
    默认情况下通过runtime.selectgo获取执行case的所有，并通过多个if语句执行对应的case中的代码

    在编译器已经对select语句进行优化后，go在运行时执行编译期间展开的runtime.selectgo函数。
    流程:随机生成一个遍历的轮训顺序pollOrder并根据channel地址生成锁定顺序lockOrder，
        根据pollOrder遍历所有case是否有可以立即处理的channel，若存在，直接获取case对应的索引并返回，
        若不存在，创建sudog结构体将当前goruntine加入到所有相关channel的收发队列并调用gopark挂起等待被唤醒
        当goruntine被唤醒时，会再次按照lockOrder遍历所有case，从中找到需要被处理的sudog索引。

## 总结:
    select需要编译器和运行时函数的通力合作




