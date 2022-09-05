import Container from "../component/Container";

interface Props {
    mdFilepath: string
}

// @ts-ignore
import bblog from "bblog"
import {BlogI} from "../component/BlogSmall";

// 用来渲染 markdown
export default function MarkDown(props: Props) {
    const blog: BlogI = bblog.getBlogDetail(props.mdFilepath)

    return <Container>
        <div className="prose dark:prose-invert"
             dangerouslySetInnerHTML={{__html: blog.content}}></div>
    </Container>

}