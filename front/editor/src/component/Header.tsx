import {FileI} from "./FileEditor";
import {UploadIcon} from "../icon";

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
    const onMenuClick = (i: MenuI) => {
        props.onMenuClick && props.onMenuClick(i)
    }

    return <section className="flex w-full bg-gray-272C38 rounded-lg  p-2 justify-center">
        {/* menu */}
        <div className="flex space-x-2 items-center">
            {
                props.menus.map(i => (
                    <div key={i.key} onClick={() => onMenuClick(i)}>menu</div>
                ))
            }
        </div>
        {/* filename */}
        <div className="flex-1 flex justify-center">{props.currFile?.name}</div>
        {/* right menu */}
        <div className="flex space-x-2 items-center">
            <button className="btn btn-ghost btn-sm" onClick={() => onMenuClick({name: "publish", key: "publish"})}>
                <UploadIcon></UploadIcon>
            </button>
        </div>
    </section>
}