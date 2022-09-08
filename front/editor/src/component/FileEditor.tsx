import {useEffect, useMemo, useRef, useState} from "react";
import CodeMirror from "@uiw/react-codemirror";
import {markdown, markdownLanguage} from '@codemirror/lang-markdown';
import {tsxLanguage, typescriptLanguage} from '@codemirror/lang-javascript';
import {basicSetup, minimalSetup} from '@uiw/codemirror-extensions-basic-setup';
import {loadLanguage, langNames, langs} from '@uiw/codemirror-extensions-langs';
import {languages} from '@codemirror/language-data';
import throttle from "lodash/throttle";
import {ReactCodeMirrorRef} from "@uiw/react-codemirror/src";

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
    onChange?: (f: FileI) => void
    onSave?: (f: FileI) => void
}

export default function FileEditor(props: Props) {
    const [body, setBody] = useState(props.file?.body || '')
    const mirror = useRef<ReactCodeMirrorRef>(null)
    const save = () => {
        props.onSave && props.onSave({
            ...props.file!,
            body
        })
    }

    const throttleChange = useMemo(() => {
        return throttle((f: FileI) => {
            props.onChange && props.onChange(f)
        }, 300)
    }, [props.onChange])

    let onChange = (e: string) => {
        console.log('editor change', props.file?.path,)
        throttleChange({
            ...props.file!,
            body: e
        })
    }

    let ext = '';
    const s = props.file?.name.split('.') || []
    if (s.length > 1) {
        ext = s[s.length - 1]
    }
    const extensions = useMemo(() => {
        switch (ext) {
            case '':
                break
            case 'md':
                return [markdown({base: markdownLanguage, codeLanguages: languages})]
            case 'yaml':
            case 'yml':
                return [langs.yaml()]
            case 'js':
            case 'jsx':
            case 'ts':
            case 'tsx':
                return [langs.tsx()]
        }

        return []
    }, [ext])
    return <div className="flex h-full flex-col overflow-y-auto" onKeyDownCapture={(e) => {
        if (e.metaKey && e.code == "KeyS") {
            e.preventDefault()
            save();
        }
    }}>
        {/* 不同的文件需要有不同的编辑器，否则历史记录会乱 */}
        <div
            className="h-full overflow-y-auto"
            key={props.file?.path}
        >
            <CodeMirror
                ref={mirror}
                autoFocus={true}
                value={props.file?.body}
                // defaultValue={props.file?.body}
                theme="dark"
                onChange={onChange}
                extensions={extensions}
            ></CodeMirror>
        </div>
        <button className="btn btn-xs mt-2" onClick={save}>save</button>
    </div>
}