import Header from "./particle/Header";
import Footer from "./particle/Footer";

interface Props {
    title: string
    page_data?: any
    logo: string
    time?: string
    children?: any
}

// @ts-ignore
import bblog from "bblog"

let params = bblog.getParams();

export default function Index(props: Props) {
    let routerBase = params.base || ''

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
        props.children
    }

    <Footer name={props.logo}></Footer>
    <div>
        {props.time}
    </div>

    </body>
    </html>
}
