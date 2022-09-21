import {useEffect, useState} from "react";
import Modal from "../component/Modal";
import "xterm/css/xterm.css"
import {Publish} from "../api/file";
import Input from "../component/Input";

interface Props {
    show: boolean,
    onConfirm?: (secret: Repo) => Promise<void>
    onClose?: () => void
    repo?: Repo
}

export interface Repo {
    remote: string
    token: string
    branch: string
}

export default function PullModal(props: Props) {
    const [repo, setRepo] = useState<Repo>(props.repo || {remote: '', token: '', branch: ''})
    useEffect(() => {
        setRepo(props.repo || {remote: '', token: '', branch: ''})
    }, [props.repo])

    const onConfirm = async () => {
        props.onConfirm && await props.onConfirm(repo)
    }

    useEffect(() => {
        console.log('pullModal')
        return () => {
            console.log('pullModal down')
        }
    }, [])

    return <Modal
        value={props.show}
        confirmBtn={"Pull"}
        title={"Pull Source"}
        onConfirm={onConfirm}
        onClose={props.onClose}
        closeBtn={"Cancel"}
        confirmClassName="btn-info"
    >
        <div>
            <Input
                label={'Remote'}
                autoFocus={true}
                type="text"
                value={repo.remote}
                onChange={(e) => {
                    setRepo({...repo, remote: e.currentTarget.value})
                }}/>
            <Input
                label={'Branch'}
                type="text"
                value={repo.branch}
                onChange={(e) => {
                    setRepo({...repo, branch: e.currentTarget.value})
                }}/>
            <Input
                label={'Token'}
                type="password"
                value={repo.token}
                onChange={(e) => {
                    setRepo({...repo, token: e.currentTarget.value})
                }}/>
        </div>

    </Modal>
}