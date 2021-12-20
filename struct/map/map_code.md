##map源码分析以及总结
##源码文件：runtime/map.go
##分析：map中有大量类似但又冗余的函数，原因之一便是没有泛型
     时间复杂度O(1)
##初始化：makemap
     map中bucket的初始化 makeBucketArray，分配bucket和溢出bucket的内存
     overflow 的定义为哈希冲突的值，用链表法解决从冲突
     底层调用makemap函数，计算得到合适的B装载因子，map容量最多可容纳装载因子乘以2^B个元素，再多就需要扩容了，如果B大于4，就需要多预留一些buckets。
     kv都存放在bmap中,bmap为最小粒度挂载，一个bmap可以放8个kv,bmap 就是我们常说的“桶”，桶里面会最多装 8 个 key
     当map的kv都不是指针，并且size都小于128字节情况下会把bmap标记为不含指针来避免gc扫描整个map
     bmap有一个overflow指针类型的字段破坏了bmap不含指针的设想，这事会把overflow移动到extra字段上
##读取：mapaccess
     mapaccess1_fat 与 mapaccess2_fat 分别对应1个与2个返回值
     以下为两种总结，都是个人总结
     a.查找过程，根据key值算出哈希值，取哈希值低位与hmap.B取模确定bucket位置，取哈希值高8位在tophash数组中查询，如果tophash[i]中存储值也哈希值相等，
     则去找该bucket中key值进行比较。如果bucket没有找到，则继续下一个overflow的bucket中查找，如果当前正处于搬迁状态，则优先从oldbuckets查找。
     如果找不到，也不会返回空值，返回相应类型的零值。
     b.hash分为低位与高位两部分，先通过低位快速找到bucket，再通过高位进一步查找，对后对比具体的key
     主要对key进行hash计算，计算后用低位的5位即末尾5位找到第几号桶，
     在通过高8位hash找到在bucket中存储的位置，当前bmap中bucket未找到则查询overflow bucket。
     若找到对应位置有数据则对比完整哈希值并返回，如果所有bucket都没有找到。则返回零值。
     如果当前map处于数据搬移状态，则优先从oldbuckets查找

##赋值：mapassign
    不支持并发操作 h.flags&hashWriting
    key的alg算法包括两种，一个是equal，用于对比；另一个是hash，也就是哈希函数
    位操作 h.flags ^= hashWriting 与 h.flags &^= hashWriting 
    根据hash找到bucket，遍历其链表下的8个bucket，对比hashtop值；如果key不在map中，判断是否需要扩容

    插入元素过程，根据key值算出哈希值，取哈希值低位与hmap.B取模确定bucket位置。
    查找该key是否存在，如果存在则直接更新值，如果没有，将key与值插入。
    插入时候如果有空位，则在第一个空位处插入，如果没空位，则增加一个溢出bucket，在溢出bucket中插入。插入过程可能会发生扩容操作。
    如果需要扩容，则会把bucket全部迁移到新申请的buckets空间中，同时多扩容一个bucket
    赋值的最后一步实际上是编译器额外生成的汇编指令来完成的，编译器和 runtime 配合，才能完成一些复杂的工作。


##扩容：hashGrow
    扩容时，会将原来的 buckets 搬运到 oldbuckets
    双倍扩容：扩容采取了一种称为“渐进式”的方式，原有的 key 并不会一次性搬迁完毕，
    每次最多只会搬迁 2 个 bucket。
    等量扩容：重新排列，极端情况下，重新排列也解决不了，map 存储就会蜕变成链表，
    性能大大降低，此时哈希因子 hash0 的设置，可以降低此类极端场景的发生。
##删除：mapdelete
    将命中的 bucket 从 oldbuckets 顺⼿搬运到buckets 中，顺便再多搬运⼀个 bucket
    引入了emptyOne与emptyRest，后者是为了加速查找
    删除的时候，会依次遍历改变top值

##map的遍历。
    首先初始化一个迭代器，因为本来就是无序的，通过随机函数算出一个bucket和遍历起始位置开始遍历。
    遍历的时候其中的overflow，那也按照这样的操作进行遍历完。最终所有的bucket遍历出来。

##.缺陷，
    a.已经扩容的map无法收缩
    b.保证并发安全是要手动读写锁，易出错，多核下的表现差
    c.难以使用sync.Pool进行重用

参考文档：
- [golang map 从源码分析实现原理（go 1.14）](https://juejin.cn/post/6888333274524221447)
- [golang map源码分析](https://www.jianshu.com/p/0d07eb2d3598)
- [年度最佳【golang】map详解](https://segmentfault.com/a/1190000023879178)
- [Golang map实践以及实现原理](https://louyuting.blog.csdn.net/article/details/99699350)
