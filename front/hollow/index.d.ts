export interface Content {
    name: string
    getContent: (getContentOpt?: getContentOpt) => string
    meta?: Record<string, any>
    ext?: string // file extension
    content?: string
    is_dir: boolean
    children?: Content[]
    toc?: TocItems[]
}

export interface TocItems {
    title: string
    items?: TocItems[]
    id: string
}

export interface getContentOpt {
    pure?: boolean // return plain text instead of rich text
}

type Config = Record<string, any>

interface GetArticlesOptions {
    sort?: (a: Content, b: Content) => boolean
    filter?: (a: Content) => boolean
    page?: number
    size?: number
    tree?: boolean // return article tree if true
}

export interface ArticleList {
    total: number
    list: Content[]
}

export function getContents(path: string, option?: GetArticlesOptions): ArticleList;

export function getConfig(): Config;

export function getContentDetail(path: string): Content;

interface MdOption {
    unwrap: boolean
}

export function md(src: string, opt?: MdOption): string;

import Hollow = require('.');

export default Hollow