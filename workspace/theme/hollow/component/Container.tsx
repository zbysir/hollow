interface Props {
    children?: any[]
}

export default function Container(props: Props) {
    return <div className="w-full max-w-6xl mx-auto
    px-5 py-6 sm:py-8 md:py-12">
        {props.children}
    </div>
}
