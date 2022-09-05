---
title: "Nil Channel 的作用"
date: 2020-02-12
tags: [Golang]
desc: 思考来源于一次忘记初始化channel，造成程序卡死。
---

思考来源于一次忘记初始化channel，造成程序卡死。

查找资料发现操作nil channels会永远阻塞，这个特性看起来似乎没什么用，甚至可能造成BUG，但Golang为何要保留这个特性？

Google之：关键字 golang nil channel

相关讨论还挺多

甚至还有新鲜的关于Go 2的提议：[proposal: language: Go 2: panic on send/receive on nil channel](https://github.com/golang/go/issues/21069)

以及Nil Chan的用法：

- [nil-channels-always-block](https://www.godesignpatterns.com/2014/05/nil-channels-always-block.html)
- [why-are-there-nil-channels-in-go](https://medium.com/justforfunc/why-are-there-nil-channels-in-go-9877cc0b2308)

目前来说，nil chan只会用在select块中，用来关闭分支。
