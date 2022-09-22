import React, {ReactFragment, ReactNode, useCallback, useEffect, useState} from "react";

interface Props {
    title: string
    className?: string,
    value: boolean
    onClose?: () => void
    onConfirm?: () => void
    onShow?: () => void
    closeBtn?: string
    closeBtnWarn?: boolean
    confirmBtn?: string
    confirmBtnWarn?: boolean
    confirmClassName?: string
    children: ReactNode,
    keyEnter?: boolean
    onKeyEnter?: () => void
    buttons?: ReactNode
}

export default function Modal(props: Props) {
    const [show, setShow] = useState(false)
    const [create, setCreate] = useState(false)
    // defer for animate
    useEffect(() => {
        setCreate(props.value)

        // 先生成 dom，再动画
        setTimeout(function () {
            setShow(props.value)
        }, 16*4)
    }, [props.value])

    useEffect(() => {
        if (show) {
            props.onShow && props.onShow()
        }
    }, [show])

    const keyDown = (event: KeyboardEvent) => {
        if (event.code === 'Escape') {
            props.onClose && props.onClose()
        } else if (event.code === 'Enter') {
            if (props.keyEnter) {
                if (props.onKeyEnter) {
                    props.onKeyEnter()
                } else {
                    props.onConfirm && props.onConfirm()
                }
            }
        }
    }

    const [confirmLoading, setConfirmLoading] = useState(false)

    useEffect(() => {
        window.addEventListener("keydown", keyDown, true);
        return () => {
            window.removeEventListener("keydown", keyDown, true);
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

    if (!create) {
        return null
    }

    return <>
        <input type="checkbox"
               className="modal-toggle" checked={show}
               onChange={() => {
               }}/>
        <div className="modal">
            <div className={"modal-box rounded-md " + props.className || ''}>
                <h3 className="font-bold text-lg ">
                    {props.title}
                </h3>
                <div className="py-4">{props.children}</div>
                <div className="modal-action mt-3">

                    {
                        props.closeBtn ?
                            <label
                                className={'btn btn-sm'
                                    + (props.closeBtnWarn ? ' btn-warning' : '')
                                }
                                onClick={props.onClose}>{props.closeBtn}</label>
                            : null
                    }

                    {
                        props.confirmBtn ?
                            <label className={"btn btn-sm"
                                + (confirmLoading ? ' loading' : '')
                                + (props.confirmBtnWarn ? ' btn-warning' : '')
                                + (props.confirmClassName ? ' ' + props.confirmClassName : '')
                            }
                                   onClick={onConfirm}>{props.confirmBtn}</label> :
                            null
                    }
                    {props.buttons}
                </div>
            </div>
        </div>
    </>
}