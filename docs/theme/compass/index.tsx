import Index from "./layout/Index"

import BlogDetail from "./page/BlogDetail";

import hollow, {Article, getArticles} from "@bysir/hollow"
import {articleRoute, sortBlog} from "./utilx";
import Menu from "./particle/Menu";
import ArticlePage from "./page/Md";

const articles = getArticles('contents',
    {
        sort: sortBlog,
        page: 1,
        size: 20,
        filter: i => (i.meta.draft !== true),
        tree: true
    }
);

let params = hollow.getConfig();

// 第一个作为首页
// const first = articles.list[0]
const first = articles.list[0]

let global = {
    title: params.title,
    logo: params.logo,
    stack: params.stack,
    footer_links: params.footer_links,
}

function flatArticles(as: Article[]): Article[] {
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

export default {
    pages: [
        {
            path: '',
            component() {
                return <Index {...global}>
                    <ArticlePage filepath={params.home_page}></ArticlePage>
                </Index>
            }
        },
        ...docs.map(b => {
            let path = '/docs/' + (b === first ? '' : articleRoute(b))
            let active = {
                ...b,
                link: path,
            }

            return {
                path: path,
                component: () => {
                    let content = b.getContent()
                    let appendLink = function (b: Article): any {
                        return {
                            ...b,
                            link: '/docs/' + (b === first ? '' : articleRoute(b)),
                            children: b.children.map(appendLink)
                        }
                    }
                    return <Index {...global}>
                        <BlogDetail {...b} content={content} menu={
                            <Menu activityMenu={active} menu={articles.list.map(appendLink)}></Menu>
                        }></BlogDetail>
                    </Index>
                }
            }
        }),
    ],

    // 将 public 文件下所有内容 copy 到 dist 下
    assets: ['statics'],
}