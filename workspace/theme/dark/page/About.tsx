interface About {
    title: any
    desc: any
    content: string
}

interface Props {
    about: About[]
}

// @ts-ignore
import bblog from "bblog"

let params = bblog.getParams();

export default function About() {
    const about = params.about
    return <div>
        <section>
            <div
                class="w-full px-5 py-6 max-w-6xl mx-auto space-y-8 md:space-y-16  sm:py-8 md:py-12  dark:text-white">
                {
                    about.map(i => (
                        <div>
                            <div className="  ">
                                <h2 className="text-4xl xl:text-5xl font-bold">{i.title}</h2>
                            </div>
                            <div className="mt-8 md:mt-12">
                                <p className="text-xl lg:text-xl text-gray-300">{i.desc}</p>
                            </div>
                            <div
                                className="
                                prose dark:prose-invert
                                max-w-none prose-p:my-1 prose-ul:my-1 prose-ul:list-outside mt-6 md:mt-8"
                                dangerouslySetInnerHTML={{__html: bblog.md(i.content)}}>
                            </div>
                        </div>
                    ))
                }
            </div>
        </section>
    </div>
}