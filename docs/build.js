const esbuild = require("esbuild");
const autoprefixer = require("autoprefixer");
const tailwindcss = require('tailwindcss')
const postCssPlugin = require("@deanc/esbuild-plugin-postcss");
const fs = require("fs-extra");

let watch

if (process.env.MODE !== 'prod') {
  watch = {
    onRebuild: function (e, result) {
      if (e) {
        console.error(e.message)
      } else {
        console.log("rebuild success")
      }
    }
  };
}
esbuild
  .build({
    entryPoints: [
      "src/index.css",
    ],
    bundle: true,
    plugins: [
      postCssPlugin({
          plugins: [tailwindcss, autoprefixer],
        },
      ),
    ],
    external: ['@bysir/hollow'],
    metafile: true,
    outdir: "statics",
    minify: true,
    sourcemap: true,
    treeShaking: true,
    target: ["chrome78"],
    watch: watch,
    write: true,
  })
  .then((e) => {
    console.log("build success")
  })
  .catch((e) => console.error(e.message));
