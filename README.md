# Hollow

Hollow 是一个快速、简洁静态博客生成器。目前只服务于自己，还未准备好给大家使用。

## Feature

- **提供 Web UI 管理文章**
  - 在任何地方（如手机上）管理你的文章
  - 云端文件也支持提交到 Git 上，不怕文件丢失
- 提供工具一键部署到 Git 仓库
- 使用 Jsx 作为主题模板开发语言
- 主题开发没有限制，代码即所得
- 快速：生成 1000 篇页面只需要 2s

## Why Hollow
Hollow 提供的工具是最简单的，代码即所得，没有复杂的规则，没有什么路由、布局概念，这极大程度的降低了主题开发成本（如果你想要开发主题的话）。

同时 Hollow 也是自由的，没有归档、分类、标签概念，它们将由"主题"自己实现（有好有坏，主题会有更多可能，但写起来更复杂）。

当你想要实现更多需求的时候，最好的方式是自己开发"主题"，而不是让某个"主题"提供给你功能。

主题开发是很简单的，比如一个最简单的主题只有一个文件，他也能运行：

> 在大多数时候，复制一个已有的主题再更改更简单

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
                return <Index title={"Hi!"} body={"I am hollow"}></Index>
            },
        },
    ],
}
```

一个完整的主题 [看这里](./workspace/theme/hollow)

## Editor

hollow 支持在服务器上运行一个 Web Editor，现在我们写 blog 不用再打开笨重的编辑器了，甚至可以在手机上进行。

Editor 特性：

- 方便的上传图片等静态文件，支持上传到本地与 OSS（暂时支持七牛云）
- 可视化文件编辑器，管理逻辑和本地文件一致。
- 用来编辑 blog 源文件，提供富文本、markdown 编辑器。
- 用来少量的编辑主题代码，如修复 bug，更改配置。

editor 不能做的：

- 主题开发：由于 hollow 运行在服务端，不自带开发环境（如 node），所以需要要使用 webpack 等构建工具还是需要在本地执行，然后将构建产物上传到 hollow 中。(在 Editor 中上传文件十分简单)。
  同时由于 hollow 的代码编辑器肯定没有你熟悉的代码编辑器好用，所以在主题开发阶段建议还是选择你趁手的编辑器，完成之后再上传到远端。

## CLI

#### `hollow download`

你可以用这个命令下载远端代码到本地，进行二次开发。
在下载文件到本地时不会删除任何本地已有的文件，如果有需要清理的文件你需要手动删除。

#### `hollow upload`

你可以用这个命令上传的源代码到远端。
