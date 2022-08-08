import Link from "./Link";

export interface BlogI {
    link: string
    name: string
    description: string
    content: string
    meta: {
        featured?: boolean,
        tags?: string[] | string
        img?: string
        date: string
    }
}

export default function BlogXS({blog}: { blog: BlogI }) {
    let link = '/blogs/' + blog.name

    return <div className="flex items-start space-x-3">
        {
            blog.meta?.img ? <Link href={link} className="block">
                <img
                    className="object-cover mb-2 overflow-hidden rounded-lg shadow-sm w-40	"
                    src={blog.meta?.img}/>
            </Link> : null
        }
        <div>
            <h2 className="font-bold text-xl">
                <Link href={link}> {blog.name}</Link></h2>
            <p className="text-sm text-gray-500">{blog.description}</p>


            <div className="flex space-x-3">
                {
                    (function () {
                        let tags = blog.meta?.tags
                        if (typeof tags === "string") {
                            tags = [tags]
                        }

                        return tags.map(i => (
                            <Link href={"/tags/" + i}>
                                <div
                                    className="bg-gray-500 items-center px-1.5 py-1 leading-none rounded-full text-xs font-medium text-white ">
                                    <span>{i}</span>
                                </div>
                            </Link>
                        ))
                    })()}
            </div>
        </div>


    </div>

}