import Link from "./Link";
import {BlogI} from "./BlogSmall";

function dateFormat(date, fmt,) {
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
        ;
    }
    ;
    return fmt;
}

export default function BlogBig({blog}: { blog: BlogI }) {
    let link = '/blogs/' + blog.name
    const name = blog.meta.title || blog.name

    let date = new Date(blog.meta?.date)
    return <div className="relative group">
        <div className="
        relative
        flex flex-col w-full dark:text-gray-100
        py-2 px-2 md:py-10 md:px-5

        hover:shadow-md
        bg-opacity-50
        z-10
    ">
            <div class="leading-relaxed">
                <h1 className="text-4xl xl:text-5xl font-bold" style={{lineHeight: '1.2'}}>
                    <Link href={link}>{name} <span
                        className="dark:text-gray-400"> {blog.meta.desc ? (' - ' + blog.meta.desc) : ''}</span>
                    </Link>
                </h1>
                <p className="pt-2 text-sm font-medium ">
                    {
                        blog.meta.featured ? <span>（置顶）</span> : null
                    }
                    <span className="mx-1">{dateFormat(new Date(blog.meta?.date), "YYYY-mm-dd")}</span>
                </p>
            </div>
        </div>

        {/* bg img */}
        <div
            className="w-full h-full absolute bg-gray-800 rounded-lg opacity-0 group-hover:opacity-50 transition-opacity duration-500	 inset-0 z-0"
        >
            {
                <img
                    className="object-cover w-full h-full rounded-lg max-h-64 shadow-md sm:max-h-96"
                    src={blog.meta?.img}/>
            }

        </div>
    </div>


}