import BlogSmall from "../component/BlogSmall";
import Link from "../component/Link";
import BlogXS from "../component/BlogXS";

// 显示所有博客的页面
export default function TagPage(props) {
    let blogs = props.blogs || []
    let tags = []
    blogs.forEach(i => {
        tags = tags.concat(i.meta?.tags)
    })

    // @ts-ignore
    tags = Array.from(new Set(tags))
    let showBlogs = blogs

    if (props.selectedTag) {
        showBlogs = blogs.filter(i => i.meta?.tags?.find(i => i === props.selectedTag))
    } else {
        // all
        showBlogs = blogs
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

    return <div className="w-full px-5 py-6 max-w-6xl mx-auto space-y-5 sm:py-8 md:py-12 sm:space-y-8 md:space-y-8 ">
        <div className="flex space-x-3 justify-center	">
            {
                tags.map(i => (
                    <Link href={"/tags/" + i}>
                        <div
                            className={[i === props.selectedTag ? 'bg-purple-500' : 'bg-gray-500', "flex items-center px-3 py-1.5 leading-none rounded-full text-xs font-medium text-white inline-block"]}>
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
                        <h3 class="py-3 text-3xl font-bold">{i.date}</h3>
                        <div className="flex flex-col space-y-4">
                            {i.blogs.map(i => <BlogXS blog={i}></BlogXS>)}
                        </div>
                    </div>
                ))
            }
          
        </div>
    </div>
}