import {useEffect, useRef, useState} from "react";
import Modal from "../component/Modal";
import {Terminal} from "xterm";
import {FitAddon} from 'xterm-addon-fit';
import {AttachAddon} from 'xterm-addon-attach';
import "xterm/css/xterm.css"

interface Props {
    show: boolean,
    onClose: () => void
    onConfirm?: () => Promise<void>
    onProgress?: (p: number) => void
    onFinish?: () => void
    ws?: any
}

export interface NewFileInfo {
    isDir: boolean,
    parentPath: string,
}

// 弹出框
export default function PublishModal(props: Props) {
    const [uploading, setUploading] = useState(false)
    const [statusMsg, setStatusMsg] = useState('')
    const [term, setTerm] = useState<Terminal>()
    const boxRef = useRef<HTMLDivElement>(null)
    const onClose = () => {
        props.onClose()
    }
    // useEffect(() => {
    //     const onResize = () => {
    //         const fitAddon = new FitAddon();
    //         fitAddon.activate(term!)
    //         fitAddon.fit();
    //         console.log('11')
    //     };
    //     onResize();
    //
    //     window.addEventListener('resize', onResize);
    //
    //     return () => {
    //         window.removeEventListener('resize', onResize);
    //     };
    // }, []);
    const onConfirm = async () => {
        // 发布 API
        term?.clear()
        props.onConfirm && await props.onConfirm()
    }

    useEffect(() => {
        // if (boxRef.current?.childElementCount) {
        //     return
        // }
        let term = new Terminal({
            // cols: 100,
            convertEol: true,
        });
        term.open(boxRef.current!);

        const fitAddon = new FitAddon();
        fitAddon.activate(term!)
        fitAddon.fit();
        term.write('Hello from \x1B[1;3;31mxterm.js\x1B[0m $ ')
        // term.dispose()

        if (boxRef.current) {
            console.log('init terminal')
        }

        setTerm(term)

        return () => {
            console.log('dispose terminal')
            term.element?.remove()
            term.dispose()
        }
    }, [])

    useEffect(() => {
        // ws
        props.ws?.Register((bs: string) => {
            // console.log('xxxx', bs)
            // term?.writeln(bs)
        })
        if (props.ws) {
            const attachAddon = new AttachAddon(props.ws?.Row());
            term?.loadAddon(attachAddon);
        }

    }, [props.ws])

    const m2 = <Modal
        className="max-w-5xl"
        value={props.show}
        confirmBtn={"Publish"}
        title={"Publish"}
        onClose={onClose}
        onConfirm={onConfirm}
        keyEnter={true}
        confirmClassName="btn-info"
    >
        {/*{...getRootProps()}*/}
        <div>

            {/*<XTerm ref={termRef}*/}
            {/*       addons={['fit', 'fullscreen', 'search']}*/}
            {/*       style={{*/}
            {/*           overflow: 'hidden',*/}
            {/*           position: 'relative',*/}
            {/*           width: '100%',*/}
            {/*           height: '100%'*/}
            {/*       }}/>*/}
            {/*<textarea*/}
            {/*    disabled*/}
            {/*    className="textarea textarea-bordered w-full"*/}
            {/*    value={statusMsg}*/}
            {/*/>*/}

            <div id="terminal" ref={boxRef}></div>
        </div>

    </Modal>

    return m2
}