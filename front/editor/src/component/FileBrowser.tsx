import {FileI} from "./FileEditor";
import {KeyValuePair} from "tailwindcss/types/config";
import {MouseEvent, useEffect, useRef, useState} from "react";
import {MenuVertical} from "./MenuVertical";
import ReactDOM from "react-dom/client";
import {Menu} from "./Menu";
import {MenuI} from "./HeaderI";
import {DocumentIcon, DocumentTextIcon, FolderIcon, FolderOpenIcon} from "../icon/Index";

export interface FileTreeI extends FileI {
    items?: FileTreeI[]
}

function FileTree(props: Props) {
    const menuDom = useRef<Element>()

    useEffect(() => {
        const c = () => {
            menuDom.current?.remove()
        }
        document.addEventListener('click', c)

        return () => {
            document.removeEventListener('click', c)
        }
    })


    const doNewFile = (currFile?: FileTreeI) => {
        props.onNewFileClick && props.onNewFileClick({isDir: false, parentPath: currFile?.dir_path!})
    }

    const doNewDir = (currFile?: FileTreeI) => {
        props.onNewFileClick && props.onNewFileClick({isDir: true, parentPath: currFile?.dir_path!})
    }

    const onContextMenuClick = (m: MenuI, target: FileTreeI) => {
        switch (m.key) {
            case 'new file':
                doNewFile(target)
                break
            case 'new directory':
                doNewDir(target)
                break
        }
    }

    const handleContextMenu = (e: MouseEvent<HTMLDivElement>) => {
        e.preventDefault()
        const dom = document.createElement('div');
        menuDom.current?.remove()

        const m = <div
            className="bg-gray-272C38 border border-gray-600 rounded-sm text-sm text-white py-1" style={
            {position: "fixed", left: e.clientX + "px", top: e.clientY + "px", minWidth: '140px', zIndex: 1000}
        }>
            <Menu menus={[
                {name: "New File", key: "new file"},
                {name: "New Directory", key: "new directory"},
                {name: "Open", key: "open"},
                {name: "Delete", key: "delete"},
            ]} onMenuClick={(m) => onContextMenuClick(m, props.tree!)}></Menu>
        </div>

        ReactDOM.createRoot(dom).render(m);

        document.body.appendChild(dom);

        menuDom.current = dom
    }

    return <div>
        <div
            onClick={() => {
                props.onFileClick && props.onFileClick(props.tree!)
            }}
            className={
                `px-1 cursor-pointer rounded-sm flex items-center
                 ${(props.currFile?.path === props.tree?.path ? 'bg-gray-600' : '')}
                 ${props.modifiedFile?.hasOwnProperty(props.tree?.path!) ? 'text-blue-600' : ''}
                 `}
            onContextMenu={handleContextMenu}
        >
            {props.tree?.is_dir ? <FolderOpenIcon></FolderOpenIcon> : <DocumentTextIcon></DocumentTextIcon>}
            <span className="ml-0.5">{props.tree?.name}</span>
        </div>
        <div className="pl-4">
            {
                props.tree?.items?.map(i => (
                    <FileTree key={i.name} {...props} tree={i}></FileTree>
                ))
            }
        </div>
    </div>
}


export interface NewFileInfo {
    isDir: boolean,
    parentPath: string
}

interface Props {
    tree?: FileTreeI
    onFileClick?: (f: FileI) => void
    onNewFileClick?: (f: NewFileInfo) => void
    currFile?: FileTreeI
    modifiedFile?: Object
}

export default function FileBrowser(props: Props) {
    const handleContextMenu = function () {
    }
    return <div className="text-sm p-2 h-full"
                onContextMenu={handleContextMenu}
    >
        <div></div>
        <FileTree {...props}></FileTree>
    </div>
}