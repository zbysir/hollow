import {routerBase} from "../config";

const base = routerBase;
// const base = '';

export default function Link(props) {
    return <a {...props} href={base + props.href}>{props.children}</a>
}