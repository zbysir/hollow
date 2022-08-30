import Container from "../component/Container";

interface Props {
    name: string,
    content: string
    meta: any
}

export default function BlogDetail(props: Props) {
    let tags = props.meta?.tags
    let name = props.meta?.title || props.name

    return <div className="container mx-auto max-w-6xl py-6 px-5">
        <div className="flex justify-center	">
            <div className="prose dark:prose-invert w-full">
                <div className="">
                    <div className="flex space-x-3">
                        {
                            tags?.map(i => (
                                <div
                                    className="bg-purple-500 flex items-center px-3 py-1.5 leading-none rounded-full text-xs font-medium text-white inline-block">
                                    <span>{i}</span>
                                </div>
                            ))
                        }
                    </div>
                </div>

                <h2> {name} </h2>
                <div dangerouslySetInnerHTML={{__html: props.content}}></div>
            </div>
        </div>
    </div>

}