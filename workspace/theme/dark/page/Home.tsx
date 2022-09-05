import BlogBig from "../component/BlogBig";
import Container from "../component/Container";

// @ts-ignore
import bblog from "bblog"
import {sortBlog} from "../utilx";

export default function Home() {
    const blogs = bblog.getBlogs('./blogs',
        {
            sort: sortBlog, page: 1, size: 20
        }
    );

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