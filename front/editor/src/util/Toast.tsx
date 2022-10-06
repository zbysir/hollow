import {EventEmitter} from 'events'
import {useCallback, useEffect, useReducer, useState} from "react";

export let EventBus = new EventEmitter();

interface ToastItem {
    text: string
    type: 'info' | 'success' | 'error'
    timeout?: number
}

export function message(type: ToastItem['type'], text: string) {
    EventBus.emit('toast', {type, text})
}

export function Toast() {
    // 这里不能使用 useStats, 是因为 setTimeout 闭包会获取到第一次 toast 的值，永远都是空
    const [toast, dispatchToast] = useReducer(function (prevState: any[], action: any) {
        let state = [...prevState]
        switch (action.type) {
            case 'add':
                state.push(action.data)
                break
            case 'remove':
                state = state.filter(i => i !== action.data)
                break
        }

        return state
    }, [])

    let listener = function (args: ToastItem) {
        dispatchToast({type: 'add', data: args})

        setTimeout(() => {
            dispatchToast({type: 'remove', data: args})
        }, args.timeout || 5000)
    };

    useEffect(() => {
        EventBus.on('toast', listener)
        return () => {
            EventBus.off('toast', listener)
        }
    }, [])

    return <div className="toast toast-top toast-end z-[1000]">
        {
            toast.map(i => (
                <div key={i.text} className="alert alert-error shadow-lg">
                    <div>
                        <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current flex-shrink-0 h-6 w-6"
                             fill="none" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
                                  d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                        </svg>
                        <span>{i.text}</span>
                    </div>
                </div>
            ))
        }

    </div>

}