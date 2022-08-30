// @ts-ignore
import bblog from "bblog"

let params = bblog.getParams();

export default function Link(props) {
    let base = params.base || ''

    return <a {...props}
              className={"text-gray-600 dark:text-white  " + props.className}
              href={base + props.href}>{props.children}</a>
}