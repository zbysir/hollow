import Home from "./page/Home";
import Header from "./particle/Header";
import Footer from "./particle/Footer";
import BlogDetail from "./page/BlogDetail";
import TagPage from "./page/TagPage";
import {routerBase} from "./config";

interface Props {
    page: 'home' | 'blog-detail' | 'tags'
    title: string
    page_data: any
    me: string
    time?: string
}

export default function Index(props: Props) {
    return <html lang="zh">
    <head>
        <meta charSet="UTF-8"/>
        <title>{props.title || 'UnTitled'}</title>
        <link href={routerBase + '/tailwind.css'} rel="stylesheet"/>
        <meta name="viewport"
              content="width=device-width, initial-scale=1.0, minimum-scale=0.5, maximum-scale=2.0, user-scalable=yes"/>
    </head>
    <body>
    <Header name={props.me}></Header>
    {
        (function () {
            switch (props.page) {
                case 'home':
                    return <Home {...props.page_data}></Home>
                case 'blog-detail':
                    return <BlogDetail {...props.page_data}></BlogDetail>
                case 'tags':
                    return <TagPage {...props.page_data}></TagPage>
            }
            return props.page
        })()
    }

    <Footer name={props.me}></Footer>
    <div>
        {props.time}
    </div>

    </body>
    </html>
}
