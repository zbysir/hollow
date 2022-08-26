import {routerBase} from "../config";

const base = routerBase;
// const base = '';

export default function Link(props) {
    return <a {...props} className={"text-gray-600 dark:text-white hover:text-gray-900 hover:dark:text-gray-200 "+ props.className} href={base + props.href}>{props.children}</a>
}