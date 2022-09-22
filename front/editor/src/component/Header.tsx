import {FileI} from "./FileEditor";
import {Menu3Icon, UploadIcon} from "../icon";
import {MenuVertical} from "./MenuVertical";
import {ReactNode} from "react";
import {FileStatus} from "../App";
import Dropdown from "./Dropdown";

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
                <Dropdown
                    menus={[
                        {
                            name: <div className={"flex space-x-2"}>
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
                                     strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                    <path strokeLinecap="round" strokeLinejoin="round"
                                          d="M7.5 7.5h-.75A2.25 2.25 0 004.5 9.75v7.5a2.25 2.25 0 002.25 2.25h7.5a2.25 2.25 0 002.25-2.25v-7.5a2.25 2.25 0 00-2.25-2.25h-.75m-6 3.75l3 3m0 0l3-3m-3 3V1.5m6 9h.75a2.25 2.25 0 012.25 2.25v7.5a2.25 2.25 0 01-2.25 2.25h-7.5a2.25 2.25 0 01-2.25-2.25v-.75"/>
                                </svg>
                                <span>Update Project</span>
                            </div>,
                            key: "update project"
                        },
                        {
                            name: <div className={"flex space-x-2"}>
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
                                     strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                    <path strokeLinecap="round" strokeLinejoin="round"
                                          d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5"/>
                                </svg>
                                <span>Push</span>
                            </div>,
                            key: "push"
                        },
                    ]}
                    position={"bottom"}
                    onMenuClick={onMenuClick}
                >
                    <button className="btn btn-ghost btn-sm">
                        <Menu3Icon></Menu3Icon>
                    </button>
                </Dropdown>
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