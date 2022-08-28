interface Props {
    children?: any[]
}

export default function Container(props: Props) {
    return <div className="w-full px-5 py-6 max-w-6xl mx-auto">
        {props.children}
    </div>
}
