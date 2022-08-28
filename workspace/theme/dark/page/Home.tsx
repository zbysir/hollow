import BlogBig from "../component/BlogBig";
import Hero from "../particle/Hero";
import BlogSmall, {BlogI} from "../component/BlogSmall";
import Container from "../component/Container";

// @ts-ignore
import bblog from "bblog"
import {sortBlog} from "../util";

export default function Home() {
    // const u = bblog.getUrl()
    const blogs = bblog.getBlog('./blogs',
        {
            sort: sortBlog, page: 1, size: 20
        }
    );
    return <div>
        <section>
            <Container>
                <div>
                    {
                        blogs.map(i => (<BlogBig blog={i}></BlogBig>))
                    }
                </div>
            </Container>
        </section>
    </div>
}