import {FileI} from "./FileEditor";
import {MenuI} from "./Header";


interface Props {
    onMenuClick?: (u: MenuI) => void
    menus: MenuI[]
    currFile?: FileI
}

export function MenuVertical(props: Props) {
    return <div className="flex flex-col">
        {
            props.menus.map(i => (
                <span key={i.key}
                    className="cursor-pointer	transform rotate-180"
                    onClick={() => props.onMenuClick && props.onMenuClick(i)}
                    style={{writingMode: "vertical-rl"}}>{i.name}</span>
            ))
        }
    </div>

}