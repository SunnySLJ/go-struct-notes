##sync.map源码分析以及总结
    map结构体包含mutex，read，dirty，misses
    read 只读数据 readOnly
    dirty 读写数据，操作 dirty 需要用 mu 进行加锁来保证并发安全
    misses 用于统计有多少次读取 read 没有命中
    readOnly.amended 用于标记 read 和 dirty 的数据是否一致

    Load读取方法，先从read读取，没有命中的话到dirty读取，同时调用missLocked方法增加misses
    如果misses大于dirty的长度，表示read和dirty数据相差太大，sync.map会将dirty数据赋值给read，而dirty会被置空
    Store写入方法，先从read修改，key不存在则去dirty修改，如果不存在则新增，同时将read中的amended标记为true表示read和dirty数据已经不一致了
    Range 会保证 read 和 dirty 是数据同步的，另一个是回调函数返回 false 会导致迭代中断,时间复杂度是O(N)
    Delete删除key对应的value,采用延迟删除的机制，首先到read查找是否存在key，若存在则执行entry.delete进行软删除，通过cas将entry.p置为nil，
    减少锁开销，若read找不到key切amended为true才会通过delete加锁硬删除。


参考文档：
- [深入浅出 Go - sync.Map 源码分析](https://xie.infoq.cn/article/ebcb070ee7fd0e273ca53b64f)

 


