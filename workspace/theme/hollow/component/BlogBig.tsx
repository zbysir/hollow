import Link from "./Link";
import {articleRoute, dateFormat} from "../utilx";
import {Article} from "@bysir/hollow";

export default function BlogBig({blog}: { blog: Article }) {
    let link = articleRoute(blog)
    const name = blog.meta.title || blog.name

    return <div className="relative group">
        <div className="
        relative
        flex flex-col w-full
        text-gray-900 dark:text-gray-100
        py-2 px-2 md:py-6 md:px-2
        text-center
        bg-opacity-50
        z-10
    ">
            <div class="leading-relaxed ">
                <h1 className="text-xl xl:text-3xl font-bold" style={{lineHeight: '1.2'}}>
                    <Link href={link}>
                        {name}
                        {/*<p className="text-base	text-gray-500 mt-2 opacity-0 hover:opacity-50">{blog.meta.desc}</p>*/}
                    </Link>
                </h1>
                <p className="pt-2 text-sm font-medium dark:text-gray-300 text-gray-700">
                    {
                        blog.meta.featured ? <span>（置顶）</span> : null
                    }
                    <span className="mx-1">{dateFormat(new Date(blog.meta?.date), "mm-dd / YY")}</span>
                </p>
            </div>
        </div>

        {/* bg img */}

        {
            blog.meta?.img ? <div
                className="w-full h-full absolute inset-0 z-0
            bg-gray-100 dark:bg-gray-800
            rounded-lg
            shadow-md
            group-hover:opacity-50 opacity-0 transition-opacity duration-500"
            >
                {
                    <img className="object-cover w-full h-full rounded-lg max-h-64 shadow-md sm:max-h-96"
                         src={blog.meta?.img}/>
                }

            </div> : null
        }

    </div>


}