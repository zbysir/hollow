---
title: 介绍下 Hollow 这个博客生成系统
slug: bblog
date: 2022-09-02
tags: [Golang, Jsx, Blog]
desc: 没有限制、规则，代码即所得。
---

https://github.com/zbysir/bblog

一款静态博客生成器，优先服务自己，还未准备好给大家使用。

## Feature

- 主题开发没有限制，没有规则，足够简单，代码即所得
- 使用 本地文件 作为数据源生成静态文件 + 实时渲染
- **支持启动 web 编辑器 在云端编辑你的文件，同时你还可以将数据源存储在一个在一个文件中（使用 blobdb），支持一键发布到 Git 仓库、支持实时渲染**
- 帮你管理静态文件，如上传到七牛，或下载文件到本地

## Why BBlog
BBlog 提供的工具是最简单的，代码即所得，没有复杂的规则，没有什么路由、布局概念，这极大程度的降低了主题开发成本（如果你想要开发主题的话）。

同时 BBlog 也是自由的，没有归档、分类、标签概念，它们将由"主题"自己实现（有好有坏，主题会有更多可能，但写起来更复杂）。

当你想要实现更多需求的时候，最好的方式是自己开发"主题"，而不是让某个"主题"提供给你功能。

比如一个最简单的项目只有一个文件：
```jsx
function Index(props) {
  return <html lang="zh" class="dark">
  <head>
    <meta charSet="UTF-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{props.title || 'UnTitled'}</title>
  </head>
  <body className="">
  <div>
    {props.body}
  </div>

  </body>
  </html>
}

export default {
  pages: [
    {
      path: '',
      component: () => {
        return <Index title={"Hi!"} body={"I am bblog"}></Index>
      },
    },
  ],
}
```

一个完整的主题 [看这里](https://github.com/zbysir/bblog/tree/master/workspace/theme/dark)

## Editor

bblog 支持运行一个 Web Editor，现在我们写 blog 不用再打开编辑器了。

Editor 使用场景：

- 方便的上传图片等静态文件，支持上传到本地与 OSS（暂时支持七牛云）
- 按照项目可视化文件编辑器，管理逻辑和编辑文件一致。
- 编辑 blog 源文件，提供富文本、markdown 编辑器。
- 少量的编辑主题代码，如修复 bug，更改配置。

editor 不能做的：

- 主题开发：由于 bblog 运行在服务端，不自带开发环境（如 node），所以需要要使用 webpack 等构建工具还是需要在本地执行，然后将构建产物上传到 bblog 中。
  bblog 提供 `bblog build --remote` 命令来帮助这个流程流畅运行。
  同时由于 bblog 的代码编辑器肯定没有你熟悉的代码编辑器好用，所以在主题开发阶段建议还是选择你趁手的编辑器，完成之后再上传到远端。

## CLI

#### `bblog download`

你可以用这个命令下载远端代码到本地，进行二次开发。
在下载文件到本地时不会删除任何本地已有的文件，如果有需要清理的文件你需要手动删除。

#### `bblog upload`

你可以用这个命令上传的源代码到远端。
