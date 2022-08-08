import Container from "../component/Container";

interface FriendLink {
    url: string
    name: string
    info?: string
}

interface Props {
    links: FriendLink[]
}

export default function Friend(props: Props) {
    let links = props.links

    return <Container>
        <div className="px-5">
            <ul className="flex space-y-3">
                {
                    links.map(i => (
                        <li>
                            <a className="link link-neutral" href={i.url} target="_blank">{i.name}</a>
                        </li>
                    ))
                }
            </ul>
        </div>
    </Container>

}