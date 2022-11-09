import {Content} from "@bysir/hollow";

export const defaultConfig = {
    logo: "Compass Theme",
    stack: "Hollow"
}

export const defaultContents: Content[] =
    [{
        name: "Demo",
        getContent: () => {
            return "<p>这篇文章在你新增任意文章后就会消失。</p>"
        },
        meta: {
            tags: ["demo", "hello"],
            date: '2022-01-01'
        },
        content: "",
        ext: "",
        is_dir: false,
        children: [],
    }]

