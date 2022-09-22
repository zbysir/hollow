import Link from "../component/Link";
import BlogXS from "../component/BlogXS";

import hollow, {Blog} from "@bysir/hollow"
import {sortBlog} from "../utilx";

let blogs = hollow.getArticles('./blogs', {
    sort: sortBlog, page: 1, size: 20
});

interface Props {
    selectedTag?: string
}

// 显示所有博客的页面
export default function TagPage(props: Props) {
    let tags = []
    blogs.list.forEach(i => {
        let items = i.meta?.tags;
        if (items) {
            tags = tags.concat(items)
        }
    })

    // @ts-ignore
    tags = Array.from(new Set(tags))
    let showBlogs: Blog[]

    if (props.selectedTag) {
        showBlogs = blogs.list.filter(i => i.meta?.tags?.find(i => i === props.selectedTag))
    } else {
        // all
        showBlogs = blogs.list
    }

    let byTime = {}

    showBlogs.forEach(i => {
        let date = i.meta?.date || '2022';
        let year = new Date(date).getFullYear()
        if (byTime[year]) {
            byTime[year].push(i)
        } else {
            byTime[year] = [i]
        }
    })
    let byTimes = []
    for (let byTimeKey in byTime) {
        byTimes.push({
            date: byTimeKey,
            blogs: byTime[byTimeKey]
        })
    }
    byTimes.sort((a, b) => {
        return b.date - a.date
    })

    return <div className="w-full px-5 py-6 max-w-6xl mx-auto space-y-5 sm:py-8 md:py-12 sm:space-y-8 md:space-y-8 ">
        <div className="flex flex-wrap space-x-3 justify-center -mb-3">
            {
                tags.map(i => (
                    <Link href={"/tags" + (i === props.selectedTag ? '' : ('/' + i))} className={"mb-3"}>
                        <div
                            className={[i === props.selectedTag ? 'bg-indigo-600' : 'bg-gray-500', "flex items-center px-3 py-1.5 leading-none rounded-full text-xs font-medium text-white inline-block"]}>
                            <span>{i}</span>
                        </div>
                    </Link>
                ))
            }
        </div>

        <div className="flex flex-col space-y-5">
            {
                byTimes.map(i => (
                    <div>
                        <h3 class="py-3 text-4xl xl:text-5xl font-bold dark:text-white text-center">{i.date}</h3>
                        <div className="flex flex-col space-y-4 py-3">
                            {i.blogs.map(i => <BlogXS blog={i}></BlogXS>)}
                        </div>
                    </div>
                ))
            }

        </div>
    </div>
}