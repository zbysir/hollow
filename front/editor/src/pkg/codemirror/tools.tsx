import {EditorView} from "@codemirror/view";
import Dropdown from "../../component/Dropdown";
import {MutableRefObject, RefObject, useCallback, useEffect, useState} from "react";
import {ReactCodeMirrorRef} from "@uiw/react-codemirror/src";
import {MenuI} from "../../component/Header";
import {EditorSelection, Transaction} from "@codemirror/state";
import {useDropzone} from "react-dropzone";
import {UploadFiles} from "../../api/file";

interface Props {
    view?: EditorView
    onChange: (e: string) => void
}

interface HeadBtnProps {
    view?: EditorView
    onClick: () => void
    onMenuClick: (e: MenuI) => void
}

function HeadBtn(props: HeadBtnProps) {
    const [currHead, setCurrHead] = useState(0)

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
        offsetY={6}
        onMenuClick={(e: MenuI) => {
            props.onMenuClick && props.onMenuClick(e)
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
        <button tabIndex={0} className="btn btn-ghost btn-xs btn-active btn-square ">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5}
                 stroke="currentColor" className="w-4 h-4">
                <path strokeLinecap="round" strokeLinejoin="round"
                      d="M5.25 8.25h15m-16.5 7.5h15m-1.8-13.5l-3.9 19.5m-2.1-19.5l-3.9 19.5"/>
            </svg>
        </button>
    </Dropdown>
}

interface ImgBtnProps {
    onFileUploaded: (img: { url: string, width: number, height: number }) => void
    onFileSelected: (fs: File[]) => void
}

function ImgBtn(props: ImgBtnProps) {
    // const [uploadFiles, setUploadFiles] = useState<File[]>([])

    const onDrop = useCallback((acceptedFiles: File[]) => {
        props.onFileSelected && props.onFileSelected(acceptedFiles)
        // setUploadFiles(acceptedFiles)
    }, [])
    const {getRootProps, isDragActive, getInputProps} = useDropzone({onDrop, noKeyboard: true})


    return <div {...getRootProps()}>
        <input {...getInputProps()} />
        <button tabIndex={0} className="btn btn-ghost btn-xs btn-active btn-square">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5}
                 stroke="currentColor" className="w-4 h-4">
                <path strokeLinecap="round" strokeLinejoin="round"
                      d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909m-18 3.75h16.5a1.5 1.5 0 001.5-1.5V6a1.5 1.5 0 00-1.5-1.5H3.75A1.5 1.5 0 002.25 6v12a1.5 1.5 0 001.5 1.5zm10.5-11.25h.008v.008h-.008V8.25zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z"/>
            </svg>
        </button>
    </div>
}

export default function CodeMirrorTools(props: Props) {
    const focus = () => {
        props.view?.focus()
    }
    const uploadDir = '/statics/img'

    const uploadFiles = async (fs: File[]) => {
        const rsp = await UploadFiles({
            project_id: 0,
            files: fs,
            path: uploadDir,
            bucket: 'project',
        })
        return rsp.data
    }

    return <div className={`bg-gray-272C38 flex space-x-1 px-4 py-1
        border-t border-base-300
        shadow
        `}>
        <HeadBtn onClick={focus} view={props.view} onMenuClick={(e) => {
        }}></HeadBtn>
        <ImgBtn
            onFileUploaded={(e) => {
            }}
            onFileSelected={async (fs) => {
                const files = await uploadFiles(fs)
                files.forEach(i => {
                    const pos = props.view?.state?.selection.main.anchor
                    const mdImg = `![](${i})`
                    let changes = {from: pos!, insert: mdImg};
                    props.view?.dispatch({
                        changes: changes,
                        userEvent: "from tools",
                        selection: EditorSelection.cursor(pos! + mdImg.length)
                    })
                })
            }}
        ></ImgBtn>
    </div>

}