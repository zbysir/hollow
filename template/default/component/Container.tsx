interface Props {
    children?: any[]
}

export default function Container(props: Props) {
    return <div className="container mx-auto max-w-6xl p-6 bg-white">
        {props.children}
    </div>
}
