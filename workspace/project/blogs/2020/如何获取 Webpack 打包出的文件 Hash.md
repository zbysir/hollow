---
title: "如何获取 Webpack 打包出的文件 Hash"
slug: get webpack hash
date: 2020-08-04
tags: [Webpack]
desc: "也许你会疑惑: 添加了hash后的文件名每次打包都会变动, 如何将最新文件名写入到页面上?"
---

## 需求
如果你使用 Vue/React脚手架搭建的项目, 你可能永远不需要这个步骤.

不过如果你需要手写webpack配置, 或者对webpack感兴趣, 也许你会疑惑: 添加了hash后的文件名每次打包都会变动, 如何将最新文件名写入到页面上?

答案是通过[html-webpack-plugin](https://github.com/jantimon/html-webpack-plugin)插件将打包之后的文件名Inject到index.html中.

此逻辑仅仅需要配置几行代码
```
new HtmlWebpackPlugin({
      filename: 'index.html',
      template: 'index.html',
      inject: true,
    })
```

很简单, 所以我们一般不用花时间研究它.

不过我在一个项目中, index.html不再由前端定义, 而是由服务端渲染输出, 所以HtmlWebpackPlugin这套逻辑不再走得通.

现在 后端就需要知晓打包后的文件名, 才能正确输出文件名到index.html中.

## 方案

#### 读取output文件夹中的文件名

这个方案最显而易见与简单, 不过它却有一些缺点:

output中文件有很多, 包括Entry和异步Chunk, 由于我们无法区分Entry和其他文件, 只能手动指定要引入的Entry文件, 代码会像这样:
```
<script :src="files[entry-a.js]"></script>
<script :src="files[entry-b.js]"></script>
<script :src="files[entry-c.js]"></script>
```

并不优雅, 我想要一个循环搞定

#### 编写Webpack插件导出文件名
如果我们要区分Entry和其他文件, 那么就只能从Webpack入手, 由于Webpack太强大(复杂), 我们需要在网上找找资料, 搜索关键字: `get webpack hash`.

- [How to inject Webpack build hash to application code](https://stackoverflow.com/questions/50228128/how-to-inject-webpack-build-hash-to-application-code)
- [webpack打包之 缓存](https://zhuanlan.zhihu.com/p/20801328?refer=jscss)

他们提到一个方案: 在插件中获取需要的文件名, 输入为一个清单文件.

有了这个清单文件, 后端就能读取它并注入到index.html中了.

不过它们提供的代码太简陋, 不能用于生产, 故继续查找资料来编写我所需要的插件:

- [通过 Webpack 的 compiler 对象的 Hooks 学会编写 Webpack 插件的编写](https://cloud.tencent.com/developer/article/1470720)
- [https://webpack.js.org/api/stats/](https://webpack.js.org/api/stats/)

同时在[html-webpack-plugin](https://github.com/jantimon/html-webpack-plugin)插件有相同功能: 将打包好的文件inject到index.html中. 所以也去翻了翻它的源码.

最后搬运过来的代码就是这样:

```javascript
class DumpAssetsPlugin {
  /*
  options: {
    filename: 'dist/access.json', // default: outputPath + "/assets.json"
  }
  */
  constructor(options) {
    options = options || {}
    this.options = {
      filename: options.filename || null,
    };
  }

  apply(compiler) {
    compiler.hooks.afterEmit.tap("ExportAssets", (compilation) => {
      // see https://webpack.js.org/api/stats/
      let stats = compilation.getStats().toJson();

      let entrypoints = compilation.entrypoints;
      let entryNames = Array.from(entrypoints.keys());

      let files = []
      for (let i = 0; i < entryNames.length; i++) {
        const entryName = entryNames[i];
        const entryFiles = entrypoints.get(entryName).getFiles();
        files.push(...entryFiles)
      }

      function unique(arr) {
        return arr.filter(function (item, index, arr) {
          return arr.indexOf(item, 0) === index;
        });
      }

      files = unique(files)

      let assets = {
        js: [],
        css: [],
        uncase: [] // 意料之外的文件
      }

      files.forEach(f => {
        const sp = f.split('.')
        const ext = sp[sp.length - 1]
        if (assets[ext]) {
          assets[ext].push(f)
        } else {
          console.warn('uncased file ext:', f)
          assets.uncase.push(f)
        }
      })

      let filename = this.options.filename
      if (!filename) {
        filename = stats.outputPath + "/assets.json"
      }

      require("fs").writeFileSync(
        filename,
        JSON.stringify(assets)
      );
    });
  }
}
```

或者, 你也可以使用我上传到NPM的包: [dump-assets-webpack-plugin](https://www.npmjs.com/package/dump-assets-webpack-plugin)

使用方法如下:
```javascript
module.exports = {
  entry: {
    index: ['./src/index.js'],
  },
  output: {
    path: __dirname + '/dist',
    filename: 'js/[name].[chunkHash:8].js'
  }
  ...

  plugins: [
    ...
    new DumpAssetsPlugin()
  ]
}
```
