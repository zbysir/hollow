import {BlogI} from "./component/BlogSmall";

export function sortBlog(a: BlogI, b: BlogI) {
    if (a.meta.featured || b.meta.featured) {
        return (a.meta.featured ? 1 : 0) > (b.meta.featured ? 1 : 0)
    }
    return a.meta.date > b.meta.date
}