import Container from "../component/Container";
import Link from "../component/Link";
import Blog from "../component/Blog";
import Hero from "../particle/Hero";
import BlogSmall from "../component/BlogSmall";

interface BlogI {
    link: string
    name: string
    meta: Object
}

export default function Home(props) {
    return <div>
        <Hero/>

        <section class="bg-white">
            <div class="w-full px-5 py-6 max-w-6xl mx-auto space-y-5 sm:py-8 md:py-12 sm:space-y-8 md:space-y-16 ">
                {
                    props.blogs.filter(i => i.meta?.featured).map(i => (<Blog blog={i}></Blog>))
                }

                <div class="flex grid grid-cols-12 pb-10 sm:px-5 gap-x-8 gap-y-16">
                    {
                        props.blogs.map(i => <BlogSmall blog={i}></BlogSmall>)
                    }

                </div>
            </div>
        </section>
    </div>
}