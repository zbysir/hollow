import {useEffect, useRef, useState} from "react";
import Modal from "../component/Modal";
import {Terminal} from "xterm";
import {FitAddon} from 'xterm-addon-fit';
import {AttachAddon} from 'xterm-addon-attach';
import "xterm/css/xterm.css"
import {Publish} from "../api/file";
import Ws from "../util/ws";

interface Props {
    show: boolean,
    onClose: () => void
    onConfirm?: () => Promise<void>
    onFinish?: () => void
    wsKey?: string
}

export interface NewFileInfo {
    isDir: boolean,
    parentPath: string,
}

// 弹出框
export default function ProcessModal(props: Props) {
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
        // props.onConfirm && await props.onConfirm()
        let xresolve: (v: unknown) => void
        const ws = new WebSocket("ws://192.168.31.119:9091/ws/" + props.wsKey);
        ws.onclose = function () {
            xresolve(false)
        }
        const attachAddon = new AttachAddon(ws);
        term?.loadAddon(attachAddon);

        return new Promise((resolve, reject) => {
            xresolve = resolve
        })
    }

    useEffect(() => {
        let term = new Terminal({
            // cols: 100,
            convertEol: true,
        });
        term.open(boxRef.current!);

        const fitAddon = new FitAddon();
        fitAddon.activate(term!)
        fitAddon.fit();
        term.write('Hello')

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
        <div>

            <div id="terminal" ref={boxRef}></div>
        </div>

    </Modal>

    return m2
}