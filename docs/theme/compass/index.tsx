import Index from "./layout/Index"

import BlogDetail from "./page/BlogDetail";

import hollow, {Content, getContents} from "@bysir/hollow"
import {articleRoute, sortBlog} from "./utilx";
import Menu from "./particle/Menu";
import ArticlePage from "./page/Md";
import {defaultConfig, defaultContents} from "./defaultdata";
import home from "../../pages/home";

let params = hollow.getConfig() || defaultConfig;

const articles = getContents('contents',
    {
        sort: sortBlog,
        page: 1,
        size: 20,
        filter: i => (i.meta.draft !== true),
        tree: true
    }
);

if (articles.list.length === 0) {
    articles.list = defaultContents
}

// 第一个作为首页
// const first = articles.list[0]
let first = null

let global = {
    title: params?.title,
    logo: params?.logo,
    stack: params?.stack,
    footer_links: params?.footer_links,
}

function flatArticles(as: Content[]): Content[] {
    let s = []
    as.forEach(i => {
        if (!i.is_dir) {
            s.push(i)
        }

        s.push(...flatArticles(i.children))
    })
    return s
}

let docs = flatArticles(articles.list)

let homepage = {}

if (params?.home_page) {
    homepage = {
        path: '',
        component() {
            return <Index {...global}>
                <ArticlePage filepath={params?.home_page}></ArticlePage>
            </Index>
        }
    }
} else if (articles.list[0]) {
    first = articles.list[0]
    homepage = {
        path: '',
        component: () => {
            let content = first.getContent()
            let appendLink = function (b: Content): any {
                return {
                    ...b,
                    link: (b === first ? '' : ('/docs/' + articleRoute(b))),
                    children: b.children.map(appendLink)
                }
            }

            return <Index {...global}>
                <BlogDetail {...first} content={content} menu={
                    <Menu activityMenu={{link: ''}} menu={articles.list.map(appendLink)}></Menu>
                }></BlogDetail>
            </Index>
        }
    }
} else {
    homepage = {
        component: () => {
            return <Index {...global}>
                Empty
            </Index>
        }
    }
}

export default {
    pages: [
        homepage,
        ...docs.map(b => {
            let path = '/docs/' + (b === first ? '' : articleRoute(b))
            return {
                path: path,
                component: () => {
                    let content = b.getContent()
                    let appendLink = function (b: Content): any {
                        return {
                            ...b,
                            link: '/docs/' + (b === first ? '' : articleRoute(b)),
                            children: b.children.map(appendLink)
                        }
                    }
                    return <Index {...global}>
                        <BlogDetail {...b} content={content} menu={
                            <Menu activityMenu={{
                                link: path,
                            }} menu={articles.list.map(appendLink)}></Menu>
                        }></BlogDetail>
                    </Index>
                }
            }
        }),
    ],

    // 将 public 文件下所有内容 copy 到 dist 下
    assets: ['statics'],
}