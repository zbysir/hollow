import Container from "../component/Container";
import Link from "../component/Link";
import Blog from "../component/Blog";
import Hero from "../particle/Hero";

interface Blog {
    link: string
    name: string
    meta: Object
}

export default function Home(props) {
    return <div>
        <Hero/>

        <section class="bg-white">
            <div class="w-full px-5 py-6 mx-auto space-y-5 sm:py-8 md:py-12 sm:space-y-8 md:space-y-16 max-w-7xl">

                {
                    props.blogs.filter(i => i.meta?.featured).map(i => {
                        let link = '/blogs/' + i.name

                        return (
                            <div className="flex flex-col items-center sm:px-5 md:flex-row">
                                <div className="w-full md:w-1/2">
                                    {
                                        i.meta?.img ? <Link href={link} className="block">
                                            <img className="object-cover w-full h-full rounded-lg max-h-64 sm:max-h-96"
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
                        )
                    })
                }


                <div class="flex grid grid-cols-12 pb-10 sm:px-5 gap-x-8 gap-y-16">

                    {
                        props.blogs.map(i => {
                            let link = '/blogs/' + i.name
                            return <div
                                className="flex flex-col items-start col-span-12 space-y-3 sm:col-span-6 xl:col-span-4">
                                {
                                    i.img ? <Link href={link} className="block">
                                        <img
                                            className="object-cover w-full mb-2 overflow-hidden rounded-lg shadow-sm max-h-56"
                                            src="https://cdn.devdojo.com/images/may2021/fruit.jpg"/>
                                    </Link> : null
                                }

                                {
                                    (function () {
                                        let tags = i.meta?.tags
                                        if (!tags?.map) {
                                            tags = [tags]
                                        }

                                        return tags.map(i => (
                                            <div
                                                className="bg-purple-500 flex items-center px-3 py-1.5 leading-none rounded-full text-xs font-medium uppercase text-white inline-block">
                                                <span>{i}</span>
                                            </div>
                                        ))
                                    })()}
                                <h2 className="text-lg font-bold sm:text-xl md:text-2xl">
                                    <Link href={link}> {i.name}</Link></h2>
                                <p className="text-sm text-gray-500">{i.description}</p>
                                {/*<p className="pt-2 text-xs font-medium"><Link href={i.link} className="mr-1 underline">Mary*/}
                                {/*    Jane</Link> · <span className="mx-1">April 17, 2021</span> · <span*/}
                                {/*    className="mx-1 text-gray-600">3 min. read</span></p>*/}
                            </div>
                        })
                    }


                </div>
            </div>
        </section>
    </div>
}