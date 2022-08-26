interface Props {
    children?: any[]
}

export default function Container(props: Props) {
    return <div className="container mx-auto max-w-6xl py-6">
        {props.children}
    </div>
}
