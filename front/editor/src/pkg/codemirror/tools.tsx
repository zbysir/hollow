import {EditorView} from "@codemirror/view";
import Dropdown from "../../component/Dropdown";
import {MutableRefObject, RefObject, useCallback, useEffect, useState} from "react";
import {ReactCodeMirrorRef} from "@uiw/react-codemirror/src";
import {MenuI} from "../../component/Header";
import {Transaction} from "@codemirror/state";

interface Props {
    view?: EditorView
    onChange: (e: string) => void
}

interface HeadBtnProps {
    view?: EditorView
    // options: MenuI[]
    onClick: () => void
}

function HeadBtn(props: HeadBtnProps) {
    const [currHead, setCurrHead] = useState(0)

    // useEffect(() => {
    //     setCurrHead(getCurrHead(props.view))
    // }, [props.view])

    function getCurrHead(view ?: EditorView) {
        let pos = view?.state?.selection.main.anchor;
        const currLine = view?.state?.doc.lineAt(pos!)
        if (currLine?.text.startsWith("#")) {
            const ss = currLine?.text.split(" ", 2)
            if (ss.length > 1) {
                return ss[0].length
            }
        }

        return 0
    }

    return <Dropdown
        menus={[
            {name: "H1", key: "1"},
            {name: "H2", key: "2"},
            {name: "H3", key: "3"},
            {name: "H4", key: "4"},
            {name: "H5", key: "5"},
        ]}
        activeMenu={currHead + ''}
        onClick={() => {
            let currHead1 = getCurrHead(props.view);
            setCurrHead(currHead1)
            props.onClick()
        }}
        onMenuClick={(e: MenuI) => {
            const view = props.view
            let pos = view?.state?.selection.main.anchor;
            const currLine = view?.state?.doc.lineAt(pos!)

            let changes = {from: 0, to: 0, insert: ''};
            let currHead = getCurrHead(props.view);
            if (currHead) {
                if (currHead == +e.key) {
                    // delete
                    changes = {
                        from: currLine?.from!,
                        to: currLine?.from! + currHead + 1,
                        insert: ''
                    };
                } else {
                    // replace
                    changes = {
                        from: currLine?.from!,
                        to: currLine?.from! + currHead + 1,
                        insert: '#'.repeat(+e.key) + ' '
                    };
                }
            } else {
                // add
                changes = {from: currLine?.from!, to: currLine?.from!, insert: '#'.repeat(+e.key) + ' '};
            }
            console.log('change', changes, view)
            view?.dispatch({changes: changes, annotations: Transaction.userEvent.of("from tools")})

        }}
    >
        <button tabIndex={0} className="btn btn-ghost btn-xs btn-active btn-square  m-1">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5}
                 stroke="currentColor" className="w-4 h-4">
                <path strokeLinecap="round" strokeLinejoin="round"
                      d="M5.25 8.25h15m-16.5 7.5h15m-1.8-13.5l-3.9 19.5m-2.1-19.5l-3.9 19.5"/>
            </svg>
        </button>
    </Dropdown>
}

export default function CodeMirrorTools(props: Props) {
    const focus = () => {
        props.view?.focus()
    }

    return <div className={"flex space-x-2 px-2"}>
        <HeadBtn onClick={focus} view={props.view}></HeadBtn>
    </div>

}