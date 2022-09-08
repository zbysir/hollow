import {FileI} from "./FileEditor";
import {Menu3Icon, UploadIcon} from "../icon";
import {MenuVertical} from "./MenuVertical";
import {ReactNode} from "react";
import {FileStatus} from "../App";

export interface MenuI {
    key: string
    name: string | ReactNode
}

interface Props {
    onMenuClick?: (u: MenuI) => void
    menus: MenuI[]
    currFile?: FileI
    drawer?: boolean
    fileStatus: FileStatus
}

export function Header(props: Props) {
    const onMenuClick = (i: MenuI) => {
        props.onMenuClick && props.onMenuClick(i)
    }

    return <section className="flex w-full space-x-2 bg-gray-272C38">
        <div className="flex flex-1 space-x-2 ">
            {/* menu */}
            <div className="flex space-x-2 items-center">
                <button className="btn btn-ghost btn-sm" onClick={() => onMenuClick({name: "menu", key: "menu"})}>
                    <Menu3Icon></Menu3Icon>
                </button>
            </div>
            {/* filename */}
            <div className="flex-1 flex justify-center nowrap items-center text-sm">{
                (props.fileStatus.modifiedFiles.find(i => i.path === props.currFile?.path) ? '*' : '') + props.currFile?.name
            }</div>
            {/* right menu */}
            <div className="flex space-x-2 items-center">
                <button className="btn btn-ghost btn-sm" onClick={() => onMenuClick({name: "publish", key: "publish"})}>
                    <UploadIcon></UploadIcon>
                </button>
            </div>
        </div>
    </section>
}