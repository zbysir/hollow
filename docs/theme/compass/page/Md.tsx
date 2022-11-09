import Container from "../component/Container";

interface Props {
    filepath: string
}

import hollow from "@bysir/hollow"

// 用来渲染 markdown
export default function ArticlePage(props: Props) {
    const content = hollow.getContentDetail(props.filepath)

    return <Container>
        <div className="prose dark:prose-invert" style={{maxWidth: '100%'}}
             dangerouslySetInnerHTML={{__html: content.content}}></div>
    </Container>
}