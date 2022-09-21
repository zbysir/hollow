import React, {useEffect, useMemo, useRef, useState} from "react";
import CodeMirror from "@uiw/react-codemirror";
import {markdown, markdownLanguage} from '@codemirror/lang-markdown';
import {EditorView, ViewUpdate} from '@codemirror/view';
import {tsxLanguage, typescriptLanguage} from '@codemirror/lang-javascript';
import {basicSetup, minimalSetup} from '@uiw/codemirror-extensions-basic-setup';
import {loadLanguage, langNames, langs} from '@uiw/codemirror-extensions-langs';
import {languages} from '@codemirror/language-data';
import throttle from "lodash/throttle";
import {ReactCodeMirrorRef} from "@uiw/react-codemirror/src";
import {oneDark} from "../pkg/codemirror/theme/one_dark";
import {Menu} from "./Menu";
import {ShowPopupMenu} from "../util/popupMenu";
import Dropdown from "./Dropdown";
import CodeMirrorTools from "../pkg/codemirror/tools";
import {useCodeMirror} from "@uiw/react-codemirror";
import {Transaction, StateEffect, EditorState} from "@codemirror/state";
import {isolateHistory, historyField} from "@codemirror/commands";
import {GetFile} from "../api/file";

export interface FileI {
    name: string
    path: string
    dir_path: string,
    is_dir: boolean
    created_at?: number
    modify_at?: number
    body: string
}

interface Props {
    file?: FileI,
    onChange?: (f: FileI) => void
    onSave?: (f: FileI) => void
}

export default function FileEditor(props: Props) {
    const [body, setBody] = useState('')
    const editor = useRef<HTMLDivElement>(null);
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

    let onChange = (e: string, viewUpdate?: ViewUpdate) => {
        // 只有有 Transaction.userEvent 的事务才会被当做更改（如 切换文件的时候重新加载页面不算更改）
        if (viewUpdate) {
            const userEvent = viewUpdate.transactions[0].annotation(Transaction.userEvent)
            if (userEvent) {
                console.log('editor change', props.file?.path,)
                throttleChange({
                    ...props.file!,
                    body: e
                })
            }
        }
        setBody(e)
    }

    let ext = '';
    const s = props.file?.name.split('.') || []
    if (s.length > 1) {
        ext = s[s.length - 1]
    }
    const extensions = useMemo(() => {
        let e = [EditorView.lineWrapping]
        switch (ext) {
            case '':
                break
            case 'md':
                e.push(markdown({base: markdownLanguage, codeLanguages: languages}))
                break
            case 'yaml':
            case 'yml':
                e.push(langs.yaml())
                break
            case 'js':
            case 'jsx':
            case 'ts':
            case 'tsx':
                e.push(langs.tsx())
                break
        }

        return e
    }, [ext])
    const {state, view, container, setContainer} = useCodeMirror({
        container: editor.current,
        onChange,
        extensions,
        theme: oneDark,
        placeholder: props.file?.path,
    });

    historyField.init(() => {
    })

    // 当 body 改变重新更改
    useEffect(() => {
        (async function () {
            const nf = await GetFile({project_id: 0, path: props.file?.path!, bucket: 'project'})
            setBody(nf.data.body)

            const currentValue = view ? view.state.doc.toString() : '';
            if (view && nf.data.body !== currentValue) {
                view.dispatch({
                    changes: {from: 0, to: currentValue.length, insert: nf.data.body},
                    annotations: [
                        // 这次修改不记录到 history
                        // 不过下面有代码会清空 history，不加这行代码也可以
                        Transaction.addToHistory.of(false),
                    ],
                    effects: [
                        EditorView.scrollIntoView(0),
                    ],
                });
                // 清空历史记录：

                // 方案 1，重新初始化 state
                // const x = view.state.toJSON({history: historyField})
                // x.history = {}
                // const state= EditorState.fromJSON(x, undefined, {history: historyField})
                // view.setState(state)

                // 方案 2，更改 histFieldValue
                // hackier to clean history
                // https://github.com/codemirror/dev/issues/651
                // @ts-ignore
                let histFieldValue: { done: any, undone: any } = view.state.field(historyField);
                histFieldValue.done = []
                histFieldValue.undone = []
            }
        })()

    }, [props.file?.path, view])

    return <div className="flex h-full flex-col overflow-y-auto" onKeyDownCapture={(e) => {
        if (e.metaKey && e.code == "KeyS") {
            e.preventDefault()
            save();
        }
    }}>
        <div
            className="h-full overflow-y-auto text-base md:text-sm bg-gray-272C38"
        >
            <div ref={editor} className={"cm-theme"}></div>
        </div>
        <CodeMirrorTools view={view} key={props.file?.path + '1'} onChange={(e) => {
            onChange(e)
        }}/>
    </div>
}