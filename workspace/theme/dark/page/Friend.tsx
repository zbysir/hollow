import Container from "../component/Container";

interface FriendLink {
    url: string
    name: string
    info?: string
}

interface Props {
    links: FriendLink[]
}

// @ts-ignore
import bblog from "bblog"
let params = bblog.getParams();
let friendLinks = params.friend_links

export default function Friend() {
    let links = friendLinks

    return <Container>

        <div className="px-5 text-white prose dark:prose-invert">
            <ul className="flex space-y-3">
                {
                    links.map(i => (
                        <li>
                            <a className="" href={i.url} target="_blank">{i.name}</a>
                        </li>
                    ))
                }
            </ul>
        </div>
    </Container>

}