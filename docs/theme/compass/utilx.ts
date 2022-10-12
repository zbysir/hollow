import {Article} from "@bysir/hollow";

export function sortBlog(a: Article, b: Article) {
    if (a.meta?.featured || b.meta?.featured) {
        return (a.meta?.featured ? 1 : 0) > (b.meta?.featured ? 1 : 0)
    }
    // console.log('xxx', a.meta.sort)
    return a.meta.sort < b.meta.sort
}

export function articleRoute(b?: Article) {
    if (!b) {
        return ''
    }
    return (b.meta?.slug || b.name)
}

export function dateFormat(date, fmt,) {
    let ret;
    const opt = {
        "Y+": date.getFullYear().toString(),        // 年
        "m+": (date.getMonth() + 1).toString(),     // 月
        "d+": date.getDate().toString(),            // 日
        "H+": date.getHours().toString(),           // 时
        "M+": date.getMinutes().toString(),         // 分
        "S+": date.getSeconds().toString()          // 秒
        // 有其他格式化字符需求可以继续添加，必须转化成字符串
    };
    for (let k in opt) {
        ret = new RegExp("(" + k + ")").exec(fmt);
        if (ret) {
            fmt = fmt.replace(ret[1], (ret[1].length == 1) ? (opt[k]) : (opt[k].padStart(ret[1].length, "0")))
        }
    }
    return fmt;
}
