import Container from "../component/Container";
import {dateFormat} from "../utilx";

interface Props {
    name: string,
    content: string
    meta?: any
    menu: any
}

export default function BlogDetail(props: Props) {
    let tags = props.meta?.tags
    let name = props.meta?.title || props.name

    return <div className="container mx-auto max-w-6xl py-6 px-5 md:py-12">
        <div className="flex">
            <div className="w-60">
                {props.menu}
            </div>
            <div className="flex-1">
                <div className="flex justify-center	">
                    <div className="prose dark:prose-invert prose-img:rounded-lg max-w-2xl w-full">
                        <h2> {name} </h2>
                        <div className="flex flex-wrap space-x-3 mb-8">
                            {props.meta?.date ?
                                <div><span className="">{dateFormat(new Date(props.meta?.date), "mm-dd / YY")}</span>
                                </div> :
                                null
                            }
                            {
                                tags?.map(i => (
                                    <div
                                        className="flex items-center text-gray-400">
                                        <span>#{i}</span>
                                    </div>
                                ))
                            }
                        </div>

                        <div dangerouslySetInnerHTML={{__html: props.content}}></div>
                    </div>
                </div>
            </div>
        </div>


    </div>

}