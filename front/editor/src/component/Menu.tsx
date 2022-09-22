import {FileI} from "./FileEditor";
import {MenuI} from "./Header";


interface Props {
    onMenuClick?: (u: MenuI) => void
    menus: MenuI[]
    active?: string
    currFile?: FileI
}

export function Menu(props: Props) {
    return <div className="flex flex-col ">
        {
            props.menus.map(i => (
                <span
                    key={i.key}
                    className={
                        `whitespace-nowrap 
                        ${props.active == i.key ? 'bg-gray-700' : ''}
                        transition-all
                        hover:bg-gray-700 cursor-pointer px-3 py-0.5`
                    }
                    onClick={() => props.onMenuClick && props.onMenuClick(i)}
                >{i.name}</span>
            ))
        }
    </div>

}