const esbuild = require("esbuild");
const autoprefixer = require("autoprefixer");
const tailwindcss = require('tailwindcss')
const stylePlugin = require('esbuild-style-plugin')

esbuild
  .build({
    entryPoints: [
      "src/index.css",
      "src/app.js",
    ],
    bundle: true,
    plugins: [
      stylePlugin({
        postcss: {plugins: [tailwindcss, autoprefixer]},
      })
    ],
    external: ['@bysir/hollow'],
    metafile: true,
    outdir: "statics",
    minify: true,
    sourcemap: true,
    treeShaking: true,
    target: ["chrome78"],
    watch: process.env.MODE !== 'prod' ? {
      onRebuild: function (e, result) {
        if (e) {
          console.error(e.message)
        } else {
          console.log("rebuild success")
        }
      }
    } : null,
    write: true,
  })
  .then((e) => {
    console.log("build success")
  })
  .catch((e) => console.error(e.message));
