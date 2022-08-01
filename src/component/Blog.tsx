import Link from "./Link";

export default function Blog({blog: i}) {
    let link = '/blogs/' + i.name

    return <div className="flex flex-col items-center sm:px-5 md:flex-row">
        <div className="w-full md:w-1/2">
            {
                i.meta?.img ? <Link href={link} className="block">
                    <img
                        className="object-cover w-full h-full rounded-lg max-h-64 sm:max-h-96"
                         src={i.meta?.img}/>
                </Link> : null
            }

        </div>
        <div
            className="flex flex-col items-start justify-center w-full h-full py-6 mb-6 md:mb-0 md:w-1/2">
            <div
                className="flex flex-col items-start justify-center h-full space-y-3 transform md:pl-10 lg:pl-16 md:space-y-5">
                <div
                    className="bg-pink-500 flex items-center pl-2 pr-3 py-1.5 leading-none rounded-full text-xs font-medium uppercase text-white inline-block">
                    <svg className="w-3.5 h-3.5 mr-1" fill="currentColor" viewBox="0 0 20 20"
                         xmlns="http://www.w3.org/2000/svg">
                        <path
                            d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"></path>
                    </svg>
                    <span>Featured</span>
                </div>
                <h1 className="text-4xl font-bold leading-none lg:text-5xl xl:text-6xl">
                    <Link href={link}>{i.name}</Link></h1>
                <p className="pt-2 text-sm font-medium"><span
                    className="mx-1">{i.meta?.date}</span>
                </p>
            </div>
        </div>
    </div>

}