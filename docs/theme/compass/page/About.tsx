interface About {
    title: any
    desc: any
    content: string
}

interface Props {
    about: About[]
}

import hollow from "@bysir/hollow"
let params = hollow.getConfig();

export default function About() {
    const about = params.about
    return <div>
        <section>
            <div
                className="max-w-6xl mx-auto w-full
                px-5 py-6 sm:py-8 md:py-12
                space-y-8 md:space-y-12
                dark:text-white
                ">
                {
                    about.map(i => (
                        <div>
                            <div className="  ">
                                <h2 className="text-2xl xl:text-2xl font-bold">{i.title}</h2>
                            </div>
                            <div className="mt-3 md:mt-4">
                                <p className="text-lg text-gray-700 dark:text-gray-300">{i.desc}</p>
                            </div>
                            <div
                                className="
                                prose dark:prose-invert
                                max-w-none prose-p:my-1 prose-ul:my-1 prose-ul:list-outside
                                mt-3 md:mt-4"
                                dangerouslySetInnerHTML={{__html: hollow.md(i.content)}}>
                            </div>
                        </div>
                    ))
                }
            </div>
        </section>
    </div>
}