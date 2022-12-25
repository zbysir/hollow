import "./index.css"
import {Terminal} from 'xterm';
import {AttachAddon} from 'xterm-addon-attach';
import ReactDOM from 'react-dom/client';

import "node_modules/xterm/css/xterm.css"
import {useEffect, useRef, useState} from "react";

window['RenderTask'] = (props: { taskKey: string }) => {

    function Task(props: { taskKey: string }) {
        const [status, setStatus] = useState('Loading...')
        const [subTitle, setSubTitle] = useState('')
        const [loading, setLoading] = useState(true)
        const terdom = useRef()

        useEffect(() => {
            let error
            const ws = new WebSocket(`ws://${window.location.host}/_dev_/ws/${props.taskKey}`);
            ws.onerror = function (e) {
                error = e
                window['RenderError']({msg: 'WebSocket 链接错误，请检查控制台'})
            }
            ws.onclose = function (e: CloseEvent) {
                if (error) {
                    return
                }
                setStatus("Load Done")
                setLoading(false)
                setSubTitle("Page will refresh in a second")
                if (!localStorage.getItem("debug")) {
                    setTimeout(() => {
                        window.location.reload()
                    }, 1000)
                }
            }

            const term = new Terminal({
                convertEol: true,
            });
            term.open(terdom.current);

            const attachAddon = new AttachAddon(ws);
            term?.loadAddon(attachAddon);
        }, [])

        return <div className="max-w-5xl mx-auto">
            <div className="my-12">
                <div className="flex justify-center items-center">
                    <svg className="transition-all duration-1000 animate-spin -ml-1 mr-3 h-5 w-5"
                         style={loading ? {} : {width: '0px'}}
                         xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                stroke-width="4"></circle>
                        <path className="opacity-75" fill="currentColor"
                              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <h2 className="text-center text-lg">{status}</h2>
                </div>
                <p className="text-center text-sm">{subTitle}</p>
            </div>
            <div ref={terdom} className="shadow-lg rounded-md overflow-hidden"></div>
        </div>
    }

    const root = ReactDOM.createRoot(document.body);
    root.render(<Task {...props}></Task>);
}


window['RenderError'] = (props: { msg: string }) => {
    function Error({msg}: { msg: string }) {
        return <div className="max-w-5xl mx-auto">
            <div className="flex justify-center items-center my-12">
                <div className="prose w-full">
                    <code>
                        <pre> {msg} </pre>
                    </code>
                </div>
            </div>
        </div>
    }

    const root = ReactDOM.createRoot(document.body);
    root.render(<Error {...props}></Error>);
}

