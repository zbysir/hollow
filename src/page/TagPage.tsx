import BlogSmall from "../component/BlogSmall";
import Link from "../component/Link";

// selectedTag
export default function TagPage(props) {
    let blogs = props.blogs || []
    let tags = []
    blogs.forEach(i => {
        tags = tags.concat(i.meta?.tags)
    })

    tags = Array.from(new Set(tags))
    let showBlogs = blogs

    if (props.selectedTag) {
        showBlogs = blogs.filter(i => i.meta?.tags?.find(i => i === props.selectedTag))
    } else {
        showBlogs = []
    }


    return <div className="w-full px-5 py-6 max-w-6xl mx-auto space-y-5 sm:py-8 md:py-12 sm:space-y-8 md:space-y-16 ">
        <div className="flex space-x-3">
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

        <div className="flex grid grid-cols-12 pb-10 sm:px-5 gap-x-8 gap-y-16">
            {
                showBlogs.map(i => <BlogSmall blog={i}></BlogSmall>)
            }
        </div>
    </div>
}