import Index from "./Index"

import Home from "./page/Home";
import BlogDetail from "./page/BlogDetail";
import TagPage from "./page/TagPage";
import About from "./page/About";

import hollow, {getArticles} from "@bysir/hollow"
import MarkDown from "./page/Md";
import {blogRoute} from "./utilx";
import Gallery from "./page/Gallery";
let blog = getArticles('./blogs');
let params = hollow.getConfig();

let global = {
    title: params.title,
    logo: params.logo,
    stack: params.stack,
    footer_links: params.footer_links,
}

let tags = []
blog.list.forEach(i => {
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
        ...blog.list.map(b => {
            return {
                path: blogRoute(b),
                component: () => {
                    let content = b.getContent()
                    // 不能这样写，因为在 golang 中没有对应的 content 字段，不能赋值成功
                    // b.content = content
                    return <Index {...global}>
                        <BlogDetail {...b} content={content}></BlogDetail>
                    </Index>
                }
            }
        }),
        {
            path: 'tags',
            component: () => {
                return <Index {...global}>
                    <TagPage></TagPage>
                </Index>
            }
        },
        ...tags.map(tag => ({
            path: 'tags/' + tag,
            component: () => {
                return <Index {...global}>
                    <TagPage selectedTag={tag}></TagPage>
                </Index>
            }
        })),
        {
            path: 'about',
            component: () => {
                return <Index {...global}>
                    <About></About>
                </Index>
            }
        },
        {
            path: 'links',
            component: () => {
                return <Index {...global}>
                    <MarkDown mdFilepath={params.links_page}></MarkDown>
                </Index>
            }
        },
        {
            path: 'gallery',
            component: () => {
                return <Index {...global}>
                    <Gallery></Gallery>
                </Index>
            }
        },
    ],

    // 将 public 文件下所有内容 copy 到 dist 下
    assets: ['statics']
}