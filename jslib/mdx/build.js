const esbuild = require("esbuild");

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
      "index.js",
    ],
    bundle: true,
    metafile: true,
    outdir: "dist",
    // minify: true,
    // sourcemap: true,
    treeShaking: true,
    watch: watch,
    write: true,
    format: "cjs",
    target: "node10",
    platform: "node",
  })
  .then((e) => {
    console.log("build success")
  })
  .catch((e) => console.error(e.message));
