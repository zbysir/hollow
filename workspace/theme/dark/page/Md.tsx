import Container from "../component/Container";

interface Props {
    mdFilepath: string
}

import hollow from "@bysir/hollow"

// 用来渲染 markdown
export default function MarkDown(props: Props) {
    const blog = hollow.getBlogDetail(props.mdFilepath)

    return <Container>
        <div className="prose dark:prose-invert"
             dangerouslySetInnerHTML={{__html: blog.content}}></div>
    </Container>

}