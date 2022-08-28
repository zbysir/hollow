import Link from "./Link";
import {BlogI} from "./BlogSmall";

export default function BlogBig({blog}: { blog: BlogI }) {
    let link = '/blogs/' + blog.name
    const name = blog.meta.title || blog.name

    return <div className="relative group">
        <div className="
        relative
    flex flex-col items-center w-full dark:text-gray-100 py-10 hover:shadow-md
    bg-opacity-50
    z-10
    ">
            <div>
                <h1 className="text-4xl font-bold text-center leading-none lg:text-5xl xl:text-6xl">
                    <Link href={link}>{name}</Link></h1>
                <p className="pt-2 text-sm font-medium text-center">
                    <span className="mx-1">{blog.meta?.date}</span>
                </p>
            </div>
        </div>

        <div className="w-full h-full absolute bg-gray-900 rounded-lg opacity-0 group-hover:opacity-40 transition-opacity	inset-0 z-0"
             >
            {

                <img
                    className="object-cover w-full h-full rounded-lg max-h-64 shadow-md sm:max-h-96"
                    src={blog.meta?.img}/>

            }

        </div>
    </div>


}