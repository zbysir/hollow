import Container from "../component/Container";

interface FriendLink {
    url: string
    name: string
    info?: string
}

interface Props {
    links: FriendLink[]
}

import hollow from "@bysir/hollow"
let params = hollow.getConfig();
let friendLinks = params.friend_links

export default function Links() {
    let links = friendLinks

    return <Container>
        <div className="prose dark:prose-invert">
            <ul className="">
                {
                    links.map(i => (
                        <li>
                            <a href={i.url} target="_blank">{i.name}</a>
                        </li>
                    ))
                }
            </ul>
        </div>
    </Container>

}