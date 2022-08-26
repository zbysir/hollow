import BlogBig from "../component/BlogBig";
import Hero from "../particle/Hero";
import BlogSmall, {BlogI} from "../component/BlogSmall";

export default function Home(props: { blogs: BlogI[] }) {
    return <div>
        <Hero/>

        <section >
            <div class="w-full px-5 py-6 max-w-6xl mx-auto space-y-5 sm:py-8 md:py-12 sm:space-y-8 md:space-y-16 ">
                {
                    props.blogs.filter(i => i.meta?.featured).map(i => (<BlogBig blog={i}></BlogBig>))
                }

                <div class="flex grid grid-cols-12 pb-10 sm:px-5 md:gap-x-8 gap-y-8">
                    {
                        props.blogs.map(i => <BlogSmall blog={i}></BlogSmall>)
                    }
                </div>
            </div>
        </section>
    </div>
}