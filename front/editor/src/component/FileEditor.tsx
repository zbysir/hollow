import {useState} from "react";

export interface FileI {
    name: string
    path: string
    is_dir: boolean
    created_at?: number
    modify_at?: number
    body: string
}

interface Props {
    file?: FileI,
    onChange?: (body: string) => void,
}

export default function FileEditor(props: Props) {
    const [body, setBody] = useState(props.file?.body || '')
    const save = () => {
        props.onChange && props.onChange(body)
    }
    return <div className="flex h-full flex-col">
        <div
            className="flex-1 overflow-y-auto"
            contentEditable
            dangerouslySetInnerHTML={{__html: props.file?.body || ''}}
            onInput={(e) => {
                setBody(e.currentTarget.innerHTML)
            }}
        ></div>
        <button className="btn btn-xs mt-2" onClick={save}>save</button>
    </div>
}