import Link from "./Link";

export default function BlogSmall({blog: i}) {
    let link = '/blogs/' + i.name

    return <div className="flex flex-col items-start col-span-12 space-y-3 sm:col-span-6 xl:col-span-4">
        {
            i.meta?.img ? <Link href={link} className="block">
                <img
                    className="object-cover w-full mb-2 overflow-hidden rounded-lg shadow-sm max-h-56"
                    src={i.meta?.img}/>
            </Link> : null
        }

        <div className="flex space-x-3">
            {
                (function () {
                    let tags = i.meta?.tags
                    if (!tags?.map) {
                        tags = [tags]
                    }

                    return tags.map(i => (
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
            <Link href={link}> {i.name}</Link></h2>
        <p className="text-sm text-gray-500">{i.description}</p>
        {/*<p className="pt-2 text-xs font-medium"><Link href={i.link} className="mr-1 underline">Mary*/}
        {/*    Jane</Link> · <span className="mx-1">April 17, 2021</span> · <span*/}
        {/*    className="mx-1 text-gray-600">3 min. read</span></p>*/}
    </div>

}