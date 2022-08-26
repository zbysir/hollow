import BlogBig from "../component/BlogBig";
import Hero from "../particle/Hero";
import BlogSmall, {BlogI} from "../component/BlogSmall";

// @ts-ignore
import bblog from "bblog"

interface About {
    title: any
    desc: any
    contents: string[]
}

interface Props {
    about: About[]
}

export default function About(props: Props) {
    return <div>
        <section>
            <div
                class="w-full px-5 py-6 max-w-6xl mx-auto space-y-5 sm:py-8 md:py-12 sm:space-y-8 md:space-y-32 dark:text-white">
                {
                    props.about.map(i => (
                        <div>
                            <div className="flex  ">
                                <h2 className="text-6xl ">{i.title}</h2>
                            </div>
                            <div className="flex   mt-16">
                                <p className="text-2xl text-gray-300">{i.desc}</p>
                            </div>
                            <div className="flex  prose  dark:prose-invert mt-8">
                                <div className="flex flex-col">
                                    {i.contents?.map(i => (
                                        bblog.md(i)
                                    ))}
                                </div>
                            </div>
                        </div>
                    ))
                }
            </div>
        </section>
    </div>
}