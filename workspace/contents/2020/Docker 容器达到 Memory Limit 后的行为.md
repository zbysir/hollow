---
title: "Docker 容器达到 Memory Limit 后的行为"
slug: docker-memory-limit
date: 2020-03-22
tags: [Docker]
desc: 坑太多了
---

当一个容器申请使用多于整个主机可用的内存时, 内核可能会杀掉容器或者是Docker daemon(守护进程)来释放内存, 这可能会导致所有服务不可用, 为了避免这个错误, 我们应该给每个容器限制合适的内存.

> [Understand the risks of running out of memory](https://docs.docker.com/config/containers/resource_constraints/#understand-the-risks-of-running-out-of-memory)

我们可以在Docker-Compose或者Docker Stack环境中使用以下配置来限制容器的内存使用:

```
version: '3.7'
services:
  mysql:
    image: mysql:5.7
    deploy:
      resources:
        limits:
          memory: 200M
      mode: global
      restart_policy:
        condition: on-failure
        delay: 5s
```

> 本文使用3.7版本的配置文件语法和swarm模式举例, 其他环境会有些差异, 其他版本的配置文件语法可以在官方文档-[compose-file](https://docs.docker.com/compose/compose-file/)
中找到.

> 更多语法, 如限制CPU等, 可以查阅[resource_constraints](https://docs.docker.com/config/containers/resource_constraints/)

接下来我们来理解上面的配置

**limits.memory**

> The maximum amount of memory the container can use. If you set this option, the minimum allowed value is 4m (4 megabyte).

容器允许的内存最大使用量, 最小值为4M.

当容器使用了大于限制的内存时, 会发生什么, 触发程序GC还是Kill?

不幸的时, 官方文档好像没有对内存限制说明得很详细, 不过Google可以帮忙, 在下面的文章中能找到一点蛛丝马迹:

- [Understanding Docker Container Memory Limit Behavior](https://medium.com/faun/understanding-docker-container-memory-limit-behavior-41add155236c)
- [Docker Compose — Memory Limits](https://linuxhint.com/docker_compose_memory_limits/)
- [https://dzone.com/articles/why-my-java-application-is-oomkilled](https://dzone.com/articles/why-my-java-application-is-oomkilled)
- [https://github.com/kubernetes/kubernetes/issues/50632](https://github.com/kubernetes/kubernetes/issues/50632)
- [https://github.com/kubernetes/kubernetes/issues/40157](https://github.com/kubernetes/kubernetes/issues/40157)

再经过试验证明当程序使用超过limits.memory限制的内存时, 容器会被Kill (cgroup干的 [resource_management_guide/sec-memory](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/6/html/resource_management_guide/sec-memory)).

简单的, 可以使用redis容器来进行这个实验: 限制内存为10M, 再添加大量数据给redis, 然后查看容器的状态.

> [如何高效地向Redis插入大量的数据](https://www.cnblogs.com/ivictor/p/5446503.html)

实际上我们不想让容器直接被Kill, 而是让Redis触发清理逻辑, 直接Kill会导致服务在一段时间内不可用(虽然会重启).

怎么办?

各种调研后发现官方提供的其他参数都不能解决这个问题, 包括memory-reservation, kernel-memory, oom-kill-disable.

- memory-reservation: 看起来是用于swam集群下的调度逻辑- [https://medium.com/@jmarcos.cano/docker-swarm-stacks-resources-limit-f447ee74cb62](https://medium.com/@jmarcos.cano/docker-swarm-stacks-resources-limit-f447ee74cb62)
- kernel-memory: 母鸡
- oom-kill-disable: 如果开启了oom-kill-disable那么当容器达到限制内存时不会被杀死, 而是假死, 实际上更惨.

看来并不能傻瓜化的解决这个问题, 现在如果我们只想触发程序的GC, 应该怎么做?

一般来说, 程序当判定到内存不足时会有自己的GC机制, 但正如这篇文章[Understanding Docker Container Memory Limit Behavior](https://medium.com/faun/understanding-docker-container-memory-limit-behavior-41add155236c)里所说, 运行在docker容器里的程序对内存限制是不可见的, 程序还是会申请大于docker limit的内存最终引起OOM Kill.

这就需要我们额外对程序进行配置, 如 redis的maxmemory配置, java的JVM配置, 不幸的是并不是所有程序都有自带的内存限制配置, 如mysql, 这种情况下建议调低程序性能 和 保证留够的程序需要的内存.

> 这篇文章有提到如何调整mysql内存: [https://marcopeg.com/2016/dockerized-mysql-crashes-lot](https://marcopeg.com/2016/dockerized-mysql-crashes-lot)

如果你的服务器开启了Swap, 有可能还会遇到一个问题: 当容器将要达到内存限制时会变得特别慢并且磁盘IO很高(达到顶峰).

这是因为我们还忽略了一个参数: memory-swap, 当没有设置memory-swap时它的值会是memory-limit的两倍, 假如设置了limit-memory=300M, 没有设置memory-swap, 这意味着容器可以使用300M内存和300M Swap. [https://docs.docker.com/config/containers/resource_constraints/#--memory-swap-details](https://docs.docker.com/config/containers/resource_constraints/#--memory-swap-details)

值得注意的是Swap并不是无损的, 相反的, 它十分慢(使用磁盘代替内存), 我们应该禁用它.

不过compose file v3并不支持memory-swap limit 的设置, 唉.

- [Docker stack deploy with compose file (version 3) memory-swap/memory-swappiness issue](https://github.com/moby/moby/issues/33742)
- [How to replace memswap_limit in docker compose 3?](https://stackoverflow.com/questions/44325949/how-to-replace-memswap-limit-in-docker-compose-3)

无奈, 那就关闭主机的swap吧.

总结 当容器达到内存限制时会发送的事情:
- 容器被Kill并重启. 解决办法是限制程序使用的内存, 如redis配置maxmemory, 或者将mysql的配置降低.
- 如果开启了swap则还有swap的副作用: 过高的磁盘占用. 解决办法是关闭主机的swap.
