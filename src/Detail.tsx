import Container from "./component/Container";

interface Props {
    name: string,
    content: string
}

export default function Detail(props: Props) {
    return <Container>
        <div className="prose">
            <h2> {props.name} </h2>
            <div>
                {props.content}
            </div>
        </div>
    </Container>

}