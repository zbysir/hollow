# Hollow

Hollow 是一个快速、简洁静态博客生成器。目前只服务于自己，还未准备好给大家使用。

> 文档待完成

## Feature

- **提供 Web UI 管理文章**
  - 在任何地方（如手机上）管理你的文章
  - 云端文件也支持提交到 Git 上，不怕文件丢失
- 提供工具一键部署到 Git 仓库
- **使用 Jsx/Tsx 作为主题模板开发语言**
- 主题开发没有限制，代码即所得
- 快速：生成 1000 篇页面只需要 2s (虽然没什么用)

## Suitable for
 - "想要从零开发主题、网站，但不想学习框架概念" 的人
 - "需要使用 Web UI 写作" 的人

## Why Hollow
在 Hollow 的世界中，代码即所得，因为主题使用 JavaScript 驱动，它是图灵完备的，所以没必要再定义限制：如路由、布局、标签、归档等。不再拘谨于框架给你的概念，这次你自己创造。

当你想要实现更多需求的时候，最好的方式是自己开发"主题"，而不是让某个"主题"提供给你功能。

> 在大多数时候，复制一个已有的主题再更改更简单

主题只有一个入口，即 index.tsx，和一个平常的 JavaScript 项目一样，支持 import 或 require 语法，如何组织你的主题，这完全取决于你。

借助于 Jsx 语法，主题开发是很简单的，比如一个最简单的主题只有一个文件：

```jsx
// index.tsx
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

一个完整的主题例子 [看这里](https://github.com/zbysir/hollow-theme/tree/master/hollow)

## Quick Start
### Install hollow
```shell
go install github.com/zbysir/hollow
```
Or use docker (recommend): https://hub.docker.com/r/bysir/hollow

### Preview Theme
This is optional and is used to select your preferred theme

```shell
docker run -p 9400:9400 bysir/hollow:master server -t https://github.com/zbysir/hollow-theme/tree/master/hollow
```

### Start your creation
- Create a project folder, e.g. `book`
- Into `book` folder
- Create `contents` folder to store contents
- Create a content, the file name is `contents/hello.md`, the content is as follows:
  ```markdown
  ---
  title: "Hello Hollow"
  date: 2022-03-22
  ---
  # Hello Hollow
  write something here
  ```
- Now your directory structure looks like this:
  ```treeview
  ./
  └── contents/
      └── hello.md
  ```
- Preview your website
  - Run Hollow server
    ```shell
    docker run -v ${PWD}:/source -p 9400:9400 bysir/hollow:master server -t https://github.com/zbysir/hollow-theme/tree/master/hollow
    ```
  - Open a browser and visit `http://localhost:9400`

### Publish

- The following command will generate static files in `.dist` directory
  ```shell
  docker run -v ${PWD}:/source bysir/hollow:master build -o /source/.dist -t https://github.com/zbysir/hollow-theme/tree/master/hollow
  ```
  ```treeview
  ./
  ├── .dist/
  └── contents/
      └── hello.md
  ```

- Put files on github page

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
