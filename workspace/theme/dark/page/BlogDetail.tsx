import Container from "../component/Container";
import {dateFormat} from "../utilx";

interface Props {
    name: string,
    content: string
    meta: any
}

export default function BlogDetail(props: Props) {
    let tags = props.meta?.tags
    let name = props.meta?.title || props.name

    return <div className="container mx-auto max-w-6xl py-6 px-5 md:py-12 font-serif">
        <div className="flex justify-center	">
            <div className="prose dark:prose-invert max-w-2xl w-full">
                <h2> {name} </h2>
                <div className="flex flex-wrap space-x-3 mb-8">
                    <div><span className="">{dateFormat(new Date(props.meta?.date), "mm-dd / YY")}</span></div>
                    {
                        tags?.map(i => (
                            <div
                                className="bg-gray-600 flex items-center px-3 py-1.5 leading-none rounded-full text-xs font-medium text-white inline-block">
                                <span>{i}</span>
                            </div>
                        ))
                    }
                </div>

                <div dangerouslySetInnerHTML={{__html: props.content}}></div>
            </div>
        </div>
    </div>

}