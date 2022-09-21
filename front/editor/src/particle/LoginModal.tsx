import {useState} from "react";
import Modal from "../component/Modal";
import "xterm/css/xterm.css"
import {Publish} from "../api/file";
import Input from "../component/Input";

interface Props {
    show: boolean,
    onConfirm?: (secret: string) => Promise<void>
}

export default function LoginModal(props: Props) {
    const [secret, setSecret] = useState('')

    const onConfirm = async () => {
        props.onConfirm && await props.onConfirm(secret)
    }

    return <Modal
        // className="max-w-5xl"
        value={props.show}
        confirmBtn={"Login"}
        title={"Login"}
        onConfirm={onConfirm}
        confirmClassName="btn-info"
    >
        <div>
            <Input
                label={'Secret'}
                autoFocus={true}
                type="password"
                value={secret}
                onChange={(e) => {
                    setSecret(e.currentTarget.value)
                }}/>
        </div>

    </Modal>
}