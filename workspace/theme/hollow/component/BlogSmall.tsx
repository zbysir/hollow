import Link from "./Link";
import {Article} from "@bysir/hollow";

export default function BlogSmall({blog}: { blog: Article }) {
    let link = '/blogs/' + blog.name
    let name = blog.meta?.title || blog.name

    return <div className="flex flex-col items-start col-span-12 space-y-3 sm:col-span-6 xl:col-span-4">
        {
            blog.meta?.img ? <Link href={link} className="block w-full">
                <img
                    className="object-cover w-full mb-2 overflow-hidden rounded-lg shadow-md h-40"
                    src={blog.meta?.img}/>
            </Link> : null
        }

        <div className="flex space-x-3">
            {
                (function () {
                    let tags = blog.meta?.tags
                    if (typeof tags === "string") {
                        tags = [tags]
                    }

                    return tags?.map(i => (
                        <Link href={"/tags/" + i}>
                            <div
                                className="bg-gray-500 items-center px-3 py-1.5 leading-none rounded-full text-xs font-medium text-white ">
                                <span>{i}</span>
                            </div>
                        </Link>
                    ))
                })()}
        </div>
        <h2 className="text-lg font-bold sm:text-xl md:text-2xl">
            <Link href={link}> {name}</Link></h2>
        <p className="text-sm text-gray-500">{blog.meta?.desc}</p>
        {/*<p className="pt-2 text-xs font-medium"><Link href={blog.link} className="mr-1 underline">Mary*/}
        {/*    Jane</Link> · <span className="mx-1">April 17, 2021</span> · <span*/}
        {/*    className="mx-1 text-gray-600">3 min. read</span></p>*/}
    </div>

}