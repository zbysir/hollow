import Modal from "./Modal";
import ReactDOM from "react-dom/client";
import {ReactNode} from "react";
import Input from "./Input";

interface ConfirmParams {
    title: string,
    children?: ReactNode,
}

interface ConfirmResult {
    ok: boolean
}

// 弹出框
export default function Confirm(p: ConfirmParams): Promise<ConfirmResult> {
    const dom = document.createElement('div');
    let xresolve: (c: ConfirmResult) => void
    const onOk = (ok: boolean) => {
        xresolve({ok})
        dom.remove()
    }

    const m = <Modal
        value={true}
        onClose={() => {
            onOk(true)
        }}
        title={p.title}
        confirmBtn={"Cancel"}
        closeBtn={"Delete"}
        closeBtnWarn={true}
        onConfirm={() => {
            onOk(false)
        }}>
        {p.children}
    </Modal>

    ReactDOM.createRoot(dom).render(m);

    document.body.querySelector("#app")?.appendChild(dom)

    return new Promise((resolve, reject) => {
        xresolve = resolve
    })
}