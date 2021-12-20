##interface源码分析以及总结

##interface包含两个核心结构eface和iface

    eface包含结构_type,_type保存了具体的类型,type.go定义了go常见的类型
    iface包含结构itab，itab包含interfacetype为接口定义方法集。而具体的方法地址都被保存在fun数组中
    iface结构itab结构也内置了_type结构，所以编译器可以通过相同的赋值过程处理这两种类型interface   
    eface赋值过程，编译器将*face结构插入定义接口的首地址，然后数据部分值进行插入，其中有偏移量。
    iface其实就是interface，这种接口内部是有函数的，eface表示empty interface
    interfae类型推断依赖于itabTable，即一个map+lock，将匹配成功的保存进来，下次可直接查询
    eface其实就是iface的子集，为什么还需要efface？显然是为了优化

##类型转换
    普通类型转换成eface，将type.int 赋值给 eface._type
    eface 转换为普通类型，比较 _type 和 type.int 是否相等，相等则断言成功，成功转换
    普通类型转换为 iface，通过runtime._type进行类型判断并转换
    iface 转换为普通类型，比较iface.itab与其是否一致
##总结：
    使用interface主要是我在不知道用户的输入类型信息对的前提下，希望能够实现一些通用数据结构或函数。
    这时候便会将空interface{}作为函数输入或者输出参数。在其他语言中，解决这种问题一般使用泛型
