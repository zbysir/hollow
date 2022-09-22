import BlogBig from "../component/BlogBig";
import Container from "../component/Container";

import {getArticles} from "@bysir/hollow"
import {sortBlog} from "../utilx";

export default function Home() {
    const blogs = getArticles('./blogs',
        {
            sort: sortBlog,
            page: 1,
            size: 20,
        }
    ).list.filter(i => (i.meta.draft !== true));

    return <section>
        <Container>
            <div className="space-y-4">
                {
                    blogs.map(i => (<BlogBig blog={i}></BlogBig>))
                }
            </div>
        </Container>
    </section>
}