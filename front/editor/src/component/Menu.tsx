import {FileI} from "./FileEditor";
import {MenuI} from "./Header";


interface Props {
    onMenuClick?: (u: MenuI) => void
    menus: MenuI[]
    currFile?: FileI
}

export function Menu(props: Props) {
    return <div className="flex flex-col">
        {
            props.menus.map(i => (
                <span
                    key={i.key}
                    className="hover:bg-gray-700 cursor-pointer pl-3 py-0.5"
                    onClick={() => props.onMenuClick && props.onMenuClick(i)}
                >{i.name}</span>
            ))
        }
    </div>

}