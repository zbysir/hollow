export interface Article {
    name: string
    getContent: () => string
    meta?: Record<string, any>
    ext: string // file extension
    content?: string
}

type ThemeConfig = Record<string, any>

interface GetArticlesOptions {
    sort: (a: Article, b: Article) => boolean
    filter: (a: Article) => boolean
    page: number
    size: number
}

interface ArticleList {
    total: number
    list: Article[]
}

export function getArticles(path: string, option?: GetArticlesOptions): ArticleList;

export function getConfig(): ThemeConfig;

export function getArticleDetail(path: string): Article;

interface MdOption {
    unwrap: boolean
}

export function md(src: string, opt?: MdOption): string;

import Hollow = require('.');

export default Hollow