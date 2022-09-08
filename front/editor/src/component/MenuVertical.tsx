import {FileI} from "./FileEditor";
import {MenuI} from "./Header";


interface Props {
    onMenuClick?: (u: MenuI) => void
    menus: MenuI[]
    activeKey?: string
}

export function MenuVertical(props: Props) {
    return <div className="flex flex-col p-0.5 bg-gray-1A1E2A rounded-br-lg opacity-70">
        {
            props.menus.map(i => (
                <span key={i.key}
                      className={`text-sm py-2 cursor-pointer transform rotate-180 text-gray-400 hover:text-current ${props.activeKey === i.key ? 'bg-gray-272C38 text-current ' : ''}`}
                      onClick={() => props.onMenuClick && props.onMenuClick(i)}
                      style={{writingMode: "vertical-rl"}}>{i.name}</span>
            ))
        }
    </div>

}