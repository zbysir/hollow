import React, {ReactFragment, ReactNode, useCallback, useEffect, useState} from "react";
import {MenuI} from "./Header";
import {Menu} from "./Menu";

interface Props {
    value?: boolean
    children: ReactNode
    menus: MenuI[]
    activeMenu?: string
    onClick?: () => void
    onOpen?: () => void
    onMenuClick?: (u: MenuI) => void
    onClose?: () => void
    // multipleClick?: boolean
}

export default function Dropdown(props: Props) {
    const [show, setShow] = useState(false)
    // defer for animate
    useEffect(() => {
        setShow(!!props.value)
    }, [props.value])
    const escFunction = (event: KeyboardEvent) => {
        if (event.code === 'Escape') {
            props.onClose && props.onClose()
        } else {
        }
    }

    const [confirmLoading, setConfirmLoading] = useState(false)

    useEffect(() => {
        window.addEventListener("keydown", escFunction, true);

        return () => {
            window.removeEventListener("keydown", escFunction, true);
        };
    });

    setTimeout(() => {
        document.body.addEventListener('click', (e) => {
            setShow(false)
        }, {once: true})
    })

    return <>
        <div className="relative" onClick={(e) => {
            props.onClick && props.onClick()
            e.stopPropagation()
        }}>
            <div onClick={(e) => {
                setShow(!show)
            }}>
                {props.children}
            </div>

            <div className={`
                absolute bottom-[100%]
                shadow
                border border-gray-600
                bg-base-100
                rounded
                origin-bottom transition-all
                ${show ? '' : 'invisible opacity-0 scale-95'}
            `}>
                <Menu menus={props.menus} onMenuClick={props.onMenuClick} active={props.activeMenu}></Menu>
            </div>
        </div>
    </>
}