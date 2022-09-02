import Index from "./Index"

import Home from "./page/Home";
import BlogDetail from "./page/BlogDetail";
import TagPage from "./page/TagPage";
import Friend from "./page/Friend";
import About from "./page/About";

// @ts-ignore
import bblog from "bblog"

let blog = bblog.getBlog('./blogs');
let params = bblog.getParams();

let global = {
    title: params.title,
    logo: params.logo,
}

let tags = []
blog.forEach(i => {
    tags = tags.concat(i.meta?.tags)
})

// @ts-ignore
tags = Array.from(new Set(tags));

export default {
    pages: [
        {
            path: '',
            component: () => {
                return <Index {...global}>
                    <Home/>
                </Index>
            },
        },
        ...blog.map(b => ({
            path: 'blogs/' + b.name,
            component: () => {
                let content = b.getContent()
                // 不能这样写，因为在 golang 中没有对应的 content 字段，不能赋值成功
                // b.content = content
                return <Index {...global}>
                    <BlogDetail {...b} content={content}></BlogDetail>
                </Index>
            }
        })),
        {
            path: 'tags',
            component: () => {
                return <Index {...global}>
                    <TagPage></TagPage>
                </Index>
            }
        },
        ...tags.map(t => ({
            path: 'tags/' + t,
            component: () => {
                return <Index {...global}>
                    <TagPage selectedTag={t}></TagPage>
                </Index>
            }
        })),
        {
            path: 'friend',
            component: () => {
                return <Index{...global}>
                    <Friend></Friend>
                </Index>
            }
        },
        {
            path: 'about',
            component: () => {
                return <Index {...global}>
                    <About></About>
                </Index>
            }
        },
    ],

    // 将 public 文件下所有内容 copy 到 dist 下
    assets: ['public']
}