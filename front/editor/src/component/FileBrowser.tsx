import {FileI} from "./FileEditor";
import {KeyValuePair} from "tailwindcss/types/config";
import {MouseEvent, useEffect, useRef, useState} from "react";
import {MenuVertical} from "./MenuVertical";
import ReactDOM from "react-dom/client";
import {Menu} from "./Menu";

export interface FileTreeI extends FileI {
    items?: FileTreeI[]
}

function FileTree(props: Props) {
    return <div>
        <div onClick={() => {
            props.onFileClick && props.onFileClick(props.tree!)
        }}
             className={
                 `px-1 cursor-pointer rounded-sm ${(props.currFile?.path === props.tree?.path ? 'bg-gray-600' : '')}
                 ${props.modifiedFile?.hasOwnProperty(props.tree?.path!) ? 'text-blue-600' : ''}
                 `}
        >{props.tree?.name}</div>
        <div className="pl-4">
            {
                props.tree?.items?.map(i => (
                    <FileTree key={i.name} {...props} tree={i}></FileTree>
                ))
            }
        </div>
    </div>
}

interface Props {
    tree?: FileTreeI
    onFileClick?: (f: FileI) => void
    onNewFileClick?: () => void
    currFile?: FileTreeI
    modifiedFile?: Object
}

export default function FileBrowser(props: Props) {
    const menuDom = useRef<Element>()

    const handleContextMenu = (e: MouseEvent<HTMLDivElement>) => {
        e.preventDefault()
        const dom = document.createElement('div');
        menuDom.current?.remove()

        const m = <div
            className="bg-gray-272C38 border border-gray-600 rounded-md text-sm text-white py-1" style={
            {position: "fixed", left: e.clientX + "px", top: e.clientY + "px", minWidth: '140px', zIndex: 1000}
        }>
            <Menu menus={[
                {name: "New", key: "new"},
                {name: "Open", key: "open"},
                {name: "Delete", key: "delete"},
            ]}></Menu>
        </div>

        ReactDOM.createRoot(dom).render(m);

        document.body.appendChild(dom);

        menuDom.current = dom
    }
    useEffect(() => {
        const c = () => {
            menuDom.current?.remove()
        }
        document.addEventListener('click', c)

        return () => {
            document.removeEventListener('click', c)
        }
    }, [])


    return <div className="text-sm p-2 h-full" onContextMenu={handleContextMenu}>
        <div onClick={() => props.onNewFileClick && props.onNewFileClick()}> +</div>
        <FileTree {...props}></FileTree>
    </div>
}