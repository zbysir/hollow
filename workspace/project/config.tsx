export default {
    theme: "dark",

// 如果需要上传到 git 仓库，需要配置 token
    git: {
        // @ts-ignore
        token: '', // 读取环境变量
        repo: "repo"
    },

    params: {
        base: "",
        title: "Bysir 的博客",
        logo: "bysir",
        friend_links: [
            {
                url: "https://blog.ache.fun/",
                name: "ache"
            }
        ],
        about: [
            {
                title: "关于 Bslog"
            },
            {
                title: "关于 Bysir"
            }
        ]
    }
}


