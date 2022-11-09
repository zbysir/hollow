import hollow from "@bysir/hollow"

let params = hollow.getConfig();

export default function Link(props) {
    let base = params?.base || ''

    return <a {...props} href={base + props.href}>{props.children}</a>
}