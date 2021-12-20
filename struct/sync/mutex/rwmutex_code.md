##rwmutex源码分析以及总结
    rwmutex结构包含mutex,writerSem,readerSem，readerCount，readerWait
    w Mutex 互斥锁，用于实现写操作之间的互斥
    writerSem 写操作信号量，用于读操作唤醒写操作
    readerSem 读操作信号量，用于写操作唤醒读操作
    readerCount 读操作的数量，不存在写操作时从0开始计数，存在写操作时从-rwmutexMaxReaders开始计数
    readerWait 写操作等待读操作的数量

    Lock通过mutex互斥锁实现互斥，将readerCount更新为负值，表示当前有写操作；当readerCount为负数时，新的读操作会被挂起
    当前若存在正在执行读操作，写操作需要等待所有读操作执行完
    UnLock写操作释放锁，将readerCount更新为正数，将所有读操作都唤醒，互斥锁释放，允许其他写操作执行
    RLock锁住rw来进行读操作获取锁，readerCount加一，如果当前存在写操作，读操作则进入阻塞状态
    RUnLock读操作释放锁，readerCount减一，如果当前读操作阻塞了写操作，readerWait-1，当readerWait为0时，
    表示阻塞写操作的所有读操作都执行完了，唤醒写操作
    综合来讲，读操作获取锁时，若当前存在写从操作则进入阻塞状态.
    读操作释放锁时，若当前读操作阻塞了写操作，所有读都执行完了再唤醒写从操作。
##分析
    RWMutex常用于大量并发读，少量并发写的场景
    综合:读锁不能阻塞读锁
        读锁需要阻塞写锁，直到所有读锁都释放
        写锁需要阻塞读锁，直到所有写锁都释放
        写锁需要阻塞写锁

参考文档：
- [Go并发 - RWMutex源码解析](https://juejin.cn/post/6968853664718913543)
    
