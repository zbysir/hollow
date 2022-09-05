export interface BlogI {
    link: string
    name: string
    description: string
    content: string
    meta?: {
        featured?: boolean,
        tags?: string[] | string
        img?: string
        date?: string
        desc?: string
        title?: string
        slug?: string
    }
}