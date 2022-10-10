import Index from "./layout/Index"

import BlogDetail from "./page/BlogDetail";

import hollow, {Article, getArticles} from "@bysir/hollow"
import {articleRoute, sortBlog} from "./utilx";
import Menu from "./particle/Menu";

const articles = getArticles('contents',
    {
        sort: sortBlog,
        page: 1,
        size: 20,
        filter: i => (i.meta.draft !== true)
    }
);

// 第一个作为首页
const first = articles.list[0]

let params = hollow.getConfig();

let global = {
    title: params.title,
    logo: params.logo,
    stack: params.stack,
    footer_links: params.footer_links,
}

function flatArticles(as :Article[]): Article[]{
    let s = []
    as.forEach(i=>{
        if (!i.is_dir) {
            s.push(i)
        }

        s.push(...flatArticles(i.children))
    })
    return s
}

let art = flatArticles(articles.list)

export default {
    pages: [
        ...art.map(b => {
            let path = b === first ? '/' : articleRoute(b);
            let active = {
                ...b,
                link: path,
            }

            return {
                path: path,
                component: () => {
                    let content = b.getContent()
                    let appendLink = function (b: Article):any{
                        return {
                            ...b,
                            link: b === first ? '/' : articleRoute(b),
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

    // 用于得到预览某一个篇文章的地址
    articleRouter: articleRoute
}