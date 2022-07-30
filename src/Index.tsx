import Home from "./page/Home";
import Header from "./particle/Header";
import Footer from "./particle/Footer";
import BlogDetail from "./page/BlogDetail";

interface Props {
    page: 'home' | 'blog-detail'
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
        <link href="/tailwind.css" rel="stylesheet"/>
        <link href="/blog/tailwind.css" rel="stylesheet"/>
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
