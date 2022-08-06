import React, {ReactFragment, ReactNode, useCallback, useEffect, useState} from "react";

interface Props {
    title: string,
    value: boolean
    onClose: () => void
    onConfirm?: () => void
    closeBtn?: string,
    confirmBtn?: string,
    children: ReactNode,
    keyEnter?: boolean
}

export default function Modal(props: Props) {
    const escFunction = (event: KeyboardEvent) => {
        if (event.code === 'Escape') {
            props.onClose()
        } else {
            // console.log('xxxx', event.code)
        }
    }

    const [confirmLoading, setConfirmLoading] = useState(false)

    useEffect(() => {
        window.addEventListener("keydown", escFunction,);

        return () => {
            window.removeEventListener("keydown", escFunction,);
        };
    }, []);

    const onConfirm = async () => {
        if (props.onConfirm) {
            setConfirmLoading(true)
            try {
                await props.onConfirm()
            } finally {
                setConfirmLoading(false)
            }
        }
    }

    return <>
        <input type="checkbox"
               className="modal-toggle" checked={props.value}
               onChange={() => {
               }}/>
        <div className="modal" onKeyUpCapture={(e) => {
            // console.log('xxx', e.code)
            if (e.code == "Enter") {
                if (props.keyEnter) {
                    props.onConfirm && props.onConfirm()
                }
            }
        }}>
            <div className="modal-box rounded-md">
                <h3 className="font-bold text-lg">
                    {props.title}
                </h3>
                <div className="py-4">{props.children}</div>
                <div className="modal-action mt-3">
                    {
                        props.confirmBtn ?
                            <label className={"btn btn-sm " + (confirmLoading ? 'loading' : '')}
                                   onClick={onConfirm}>{props.confirmBtn}</label> :
                            null
                    }
                    <label className={'btn btn-sm'}
                           onClick={props.onClose}>{props.closeBtn || 'Cancel'}</label>
                </div>
            </div>
        </div>
    </>
}