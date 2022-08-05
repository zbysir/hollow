import {FileI} from "./FileEditor";

export interface FileTreeI extends FileI {
    items?: FileTreeI[]
}

function FileTree({tree, onFileClick}: Props) {
    return <div>
        <div onClick={() => {
            onFileClick && onFileClick(tree!)
        }}>{tree?.name}</div>
        <div className="pl-4">
            {
                tree?.items?.map(i => (
                    <FileTree key={i.name} tree={i} onFileClick={onFileClick}></FileTree>
                ))
            }
        </div>
    </div>
}

interface Props {
    tree?: FileTreeI
    onFileClick?: (f: FileI) => void
    onNewFileClick?: () => void
}

export default function FileBrowser({tree, onFileClick, onNewFileClick}: Props) {
    return <div className="text-base  ">
        <div onClick={() => onNewFileClick && onNewFileClick()}> +</div>
        <FileTree tree={tree} onFileClick={onFileClick}></FileTree>
    </div>
}