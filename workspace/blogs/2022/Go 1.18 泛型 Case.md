---
title: 学习 Go 1.18 泛型
slug: generics_in_go
date: 2022-09-22
tags: [Golang]
desc: 泛型也要慢慢用起来呀。
draft: true
---

# 语法
## 方法（Method）
Method：在结构体上的函数成员

👌

🚫

## 函数（Function）

# 实际用途
学习了语法，来看几个实际用途

## errors.As

泛型前

```go
func main() {
    var err error = MyError{}

    var s MyError
    if errors.As(err, &s) {
        print(s.InnerMsg)
    }
}
```

泛型后


> 能少写一行是一行

```go
func main() {
    var err error = MyError{}

    if s, ok := GenericsErrorAs[MyError](err); ok {
        print(s.InnerMsg)
    }
}


func GenericsErrorAs[T error](err error) (t T, ok bool) {
    ok = errors.As(err, &t)
    return
}

```
