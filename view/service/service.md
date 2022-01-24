#### 微服务架构是什么样子的

通常传统的项目体积庞大，需求、设计、开发、测试、部署流程固定。新功能需要在原项目上做修改。

但是微服务可以看做是对大项目的拆分，是在快速迭代更新上线的需求下产生的。新的功能模块会发布成新的服务组件，与其他已发布的服务组件一同协作。 服务内部有多个生产者和消费者，通常以http rest的方式调用，服务总体以一个（或几个）服务的形式呈现给客户使用。

微服务架构是一种思想对微服务架构我们没有一个明确的定义，但简单来说微服务架构是：

采用一组服务的方式来构建一个应用，服务独立部署在不同的进程中，不同服务通过一些轻量级交互机制来通信，例如 RPC、HTTP 等，服务可独立扩展伸缩，每个服务定义了明确的边界，不同的服务甚至可以采用不同的编程语言来实现，由独立的团队来维护。

Golang的微服务框架[kit](https://gokit.io/)中有详细的微服务的例子,可以参考学习.

微服务架构设计包括:

1. 服务熔断降级限流机制 熔断降级的概念(`Rate Limiter` 限流器,`Circuit breaker` 断路器).
2. 框架调用方式解耦方式 `Kit` 或 `Istio` 或 `Micro` 服务发现(consul zookeeper kubeneters etcd ) RPC调用框架.
3. 链路监控,`zipkin`和`prometheus`.
4. 多级缓存.
5. 网关 (`kong gateway`).
6. Docker部署管理 `Kubenetters`.
7. 自动集成部署 CI/CD 实践.
8. 自动扩容机制规则.
9. 压测 优化.
10. `Trasport` 数据传输(序列化和反序列化).
11. `Logging` 日志.
12. `Metrics` 指针对每个请求信息的仪表盘化.

微服务架构介绍详细的可以参考:

[Microservice Architectures](https://www.pst.ifi.lmu.de/Lehre/wise-14-15/mse/microservice-architectures.pdf)
