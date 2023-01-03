> 优先级由上到下递减

## 主要
- [x] 主题支持 git 协议远程获取
- [x] Server 静态文件 304，而不是默认的 200

## 优化
- [x] config 支持 js 脚本（部分数据结构，如列表使用 js 能更好表达）

## 完善主题

- [ ] 为 hollow 写一个操作手册主题，抄袭：https://docs.yjs.dev/ | https://nacos.io/en-us/index.html 
- [x] 实现模糊搜索，React 与 https://github.com/krisk/Fuse

## 扩展

- [x] 项目文件考虑支持 Jsx（废弃，使用 mdx 就好）
- [x] 支持 [MDX](https://mdxjs.com/packages/mdx/)（经过实验，由于兼容问题，使用 Goja 运行 mdx 项目是不容易实现的，考虑自己实现子集）
  - 两个实现方案：
    - [x] 先将 整个 markdown 转换成 jsx node（需要写插件特殊处理），然后整个文件交由 jsx 运行。这有个好处，mdx 可以做成 loader（做到 gojsx 里），支持 import A from "a.mdx"。
    - 只处理 markdown 中的 jsx block。更快的性能，但是如果要处理 {}, inline 的语法会特别麻烦。不适用。

## 发布

- [x] 将 Hollow 发布为 Docker 镜像: https://hub.docker.com/r/bysir/hollow
- [x] 将主题独立为其他仓库
- [x] GitHub Action 编译、部署源文件

## 评论
- https://github.com/gitalk/gitalk

## 优化 Web 编辑器

> 编辑器比想象中麻烦，优先级先降低

- [ ] 图片上传
    - [x] 选择图片上传
    - [ ] 粘贴
- [x] 优化文件夹打开逻辑：默认关闭，记录打开状态
- [x] 
  优化文件打开、修改交互 [codesandbox](https://codesandbox.io/s/uploadcare-react-widget-props-example-forked-g1q3z8?file=/src/index.js)
- [ ] 导入导出文件（支持 Git Clone），一般用于导入主题
- [ ] 批量上传支持过滤 gitignore 规则（js 实现有点麻烦，可以直接写黑名单，如 node_modules）

## 多项目管理

Under consideration