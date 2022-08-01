import Index from "./Index"

// @ts-ignore
import db from "db"

let blog = db.getSource('./blogs');

let global = {
    title: "bysir 的博客",
    me: "bysir",
}

let tags = []
blog.forEach(i => {
    tags = tags.concat(i.meta?.tags)
})

// @ts-ignore
export let routerBase = process.env?.base || ''

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
                let content = b.getContent()
                // 不能这样写，因为在 golang 中没有对应的 content 字段，不能赋值成功
                // b.content = content
                return Index({
                    ...global, page: 'blog-detail', page_data: {
                        ...b,
                        content,
                    }
                })
            }
        })),
        {
            name: 'tags',
            component: () => {
                return Index({...global, page: 'tags', page_data: {blogs: blog}})
            }
        },
        ...tags.map(t => ({
            name: 'tags/' + t,
            component: () => {
                return Index({...global, page: 'tags', page_data: {blogs: blog, selectedTag: t}})
            }
        }))
    ],

    // 将 public 文件下所有内容 copy 到 dist 下
    assets: ['public']
}