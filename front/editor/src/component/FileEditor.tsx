import {useEffect, useRef, useState} from "react";
import CodeMirror from "@uiw/react-codemirror";
import {tsxLanguage, typescriptLanguage} from '@codemirror/lang-javascript';
import {markdown} from '@codemirror/lang-markdown';
import {basicSetup, minimalSetup} from '@uiw/codemirror-extensions-basic-setup';
import {loadLanguage, langNames, langs} from '@uiw/codemirror-extensions-langs';

export interface FileI {
    name: string
    path: string
    dir_path: string,
    is_dir: boolean
    created_at?: number
    modify_at?: number
    body: string
    modified?: boolean,
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

    return <div className="flex h-full flex-col overflow-y-auto" onKeyDownCapture={(e) => {
        console.log('xxx')
        if (e.metaKey && e.code == "KeyS") {
            e.preventDefault()
            save();
        }
    }}>
        <div className="h-full overflow-y-auto">
            <CodeMirror
                value={props.file?.body}
                theme="dark"
                onChange={(e) => setBody(e)}
                extensions={[langs.tsx()]}
            ></CodeMirror>
        </div>
        <button className="btn btn-xs mt-2" onClick={save}>save</button>
    </div>
}