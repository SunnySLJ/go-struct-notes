##once源码分析以及总结
      once用于只执行一次的场景，常用于懒汉式单利模式
    其结构体once包含done和mutex，Do方法用atomic读取保证原子load，若发现done已经是1，则直接返回。
    若不是，则通过mutex进行加锁操作
