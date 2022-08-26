import Home from "./page/Home";
import Header from "./particle/Header";
import Footer from "./particle/Footer";
import BlogDetail from "./page/BlogDetail";
import TagPage from "./page/TagPage";
import {routerBase} from "./config";
import Friend from "./page/Friend";
import About from "./page/About";

interface Props {
    page: 'home' | 'blog-detail' | 'tags' | 'friend' | 'about'
    title: string
    page_data: any
    logo: string
    time?: string
}

export default function Index(props: Props) {
    return <html lang="zh" class="dark">
    <head>
        <meta charSet="UTF-8"/>
        <title>{props.title || 'UnTitled'}</title>
        <link href={routerBase + '/tailwind.css'} rel="stylesheet"/>
        <meta name="viewport"
              content="width=device-width, initial-scale=1.0, minimum-scale=0.5, maximum-scale=2.0, user-scalable=yes"/>
    </head>
    <body className="bg-gray-50 dark:bg-gray-800">
    <Header name={props.logo}></Header>
    {
        (function () {
            switch (props.page) {
                case 'home':
                    return <Home {...props.page_data}></Home>
                case 'blog-detail':
                    return <BlogDetail {...props.page_data}></BlogDetail>
                case 'tags':
                    return <TagPage {...props.page_data}></TagPage>
                case 'friend':
                    return <Friend {...props.page_data}></Friend>
                case 'about':
                    return <About {...props.page_data}></About>
            }
            return props.page
        })()
    }

    <Footer name={props.logo}></Footer>
    <div>
        {props.time}
    </div>

    </body>
    </html>
}
