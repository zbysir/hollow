import {FileI} from "./FileEditor";

export interface MenuI {
    key: string
    name: string
}

interface Props {
    onMenuClick?: (u: MenuI) => void
    menus: MenuI[]
    currFile?: FileI
}

export function Header(props: Props) {
    return <section className="flex w-full bg-gray-272C38 rounded-lg  p-2 justify-center">
        <div className="flex space-x-2 items-center">
            {
                props.menus.map(i => (
                    <div key={i.key} onClick={() => props.onMenuClick && props.onMenuClick(i)}>menu</div>
                ))
            }
        </div>
        <div className="flex-1 flex justify-center">{props.currFile?.name}</div>
        <div></div>
    </section>
}