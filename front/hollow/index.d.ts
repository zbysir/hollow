export interface Blog {
    name: string
    getContent: () => string
    meta?: Record<string, any>
    ext: string // file extension
    content?: string
}

type ThemeConfig = Record<string, any>

interface GetBlogsOptions {
    sort: (a: Blog, b: Blog) => boolean
    page: number
    size: number
}

interface BlogList {
    total: number
    list: Blog[]
}

export function getBlogs(path: string, option?: GetBlogsOptions): BlogList;

export function getConfig(): ThemeConfig;

export function getBlogDetail(path: string): Blog;

interface MdOption {
    unwrap: boolean
}

export function md(src: string, opt?: MdOption): string;

import Hollow = require('.');

export default Hollow