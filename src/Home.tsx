import Container from "./component/Container";

interface Blog {
    link: string
    name: string
}

export default function Home(props) {
    return <Container>
        <div className="prose">
            <h2> 最新 Blogs </h2>
            <ul className="list-disc list-inside">
                {
                    props.blogs.map(i => (
                        <li><a href={'./blogs/'+i.link}>{i.name}</a></li>
                    ))
                }
            </ul>
        </div>
    </Container>
}