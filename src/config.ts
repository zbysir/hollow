import Index from "./Index"

// @ts-ignore
import db from "db"

let blog = db.getSource('./blogs');

let global = {
    title: "bysir 的博客",
    me: "bysir",
}

export let routerBase = ''

export default {
    pages: [
        {
            name: '',
            component: function () {
                return Index({...global, page: 'home', page_data: {blogs: blog}})
            },
        },
        ...blog.map(b => ({
            name: 'blogs/' + b.name,
            component: function () {
                b.content = b.getContent()
                return Index({...global, page: 'blog-detail', page_data: b})
            }
        }))
    ],

    // 将 public 文件下所有内容 copy 到 dist 下
    assets: ['public']
}