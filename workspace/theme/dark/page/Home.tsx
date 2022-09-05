import BlogBig from "../component/BlogBig";
import Container from "../component/Container";

import {getBlogs} from "@bysir/hollow"
import {sortBlog} from "../utilx";

export default function Home() {
    const blogs = getBlogs('./blogs',
        {
            sort: sortBlog, page: 1, size: 20
        }
    ).list;

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