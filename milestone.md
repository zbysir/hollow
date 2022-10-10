## 优化 Web 编辑器
- [ ] 图片上传
  - [x] 选择图片上传
  - [ ] 粘贴
- [x] 优化文件夹打开逻辑：默认关闭，记录打开状态
- [x] 优化文件打开、修改交互 [codesandbox](https://codesandbox.io/s/uploadcare-react-widget-props-example-forked-g1q3z8?file=/src/index.js)
- [ ] 导入导出文件（支持 Git Clone），一般用于导入主题
- [ ] 批量上传支持过滤 gitignore 规则（js 实现有点麻烦，可以直接写黑名单，如 node_modules）

## 完善主题
- [ ] 为 hollow 写一个操作手册主题，抄袭：https://docs.yjs.dev/ (doing)
- [ ] 实现模糊搜索，React 与 https://github.com/krisk/Fuse

## 发布
- [x] 将 Hollow 发布为 Docker 镜像: https://hub.docker.com/r/bysir/hollow
- [ ] GitHub CI 编译、部署源文件

## 多项目管理
Under consideration