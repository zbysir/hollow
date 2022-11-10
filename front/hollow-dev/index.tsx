import "./index.css"
import {Terminal} from 'xterm';
import {AttachAddon} from 'xterm-addon-attach';

import "node_modules/xterm/css/xterm.css"

window['RenderTask'] = ({taskKey}: { taskKey: string }) => {
    const box = document.createElement("div")
    box.className = " max-w-5xl mx-auto"
    box.innerHTML = '<div class="flex justify-center items-center my-12">' +
        '<svg id="loading-icon" class="transition-all duration-1000 animate-spin -ml-1 mr-3 h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">\n' +
        '        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>\n' +
        '        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>\n' +
        '      </svg>' +
        '<h2 id="status" class="text-center text-lg"></h2>' +
        '</div>' +
        '<div id="terminal" class="shadow-lg rounded-md overflow-hidden"></div>'
    document.body.append(box)
    const status = document.getElementById('status')

    function setStatus(s: string, done: boolean) {
        status.innerHTML = s
        if (done) {
            document.getElementById('loading-icon').style.width = '0px'
        }
    }

    setStatus("Loading...", false)

    const ws = new WebSocket("ws://localhost:8083/_dev_/ws/" + taskKey);
    ws.onclose = function () {
        setStatus("Done, Refresh page after a second", true)
        setTimeout(() => {
            window.location.reload()
        }, 1000)
    }

    const term = new Terminal({
        convertEol: true,
    });
    term.open(document.getElementById('terminal'));

    const attachAddon = new AttachAddon(ws);
    term?.loadAddon(attachAddon);
}


window['RenderError'] = ({msg}: { msg: string }) => {
    const box = document.createElement("div")
    box.className = " max-w-5xl mx-auto"
    box.innerHTML = '<div class="flex justify-center items-center my-12">' +
        '<div class="prose w-full">' +
        '<code> <pre>' +
        msg +
        '</pre></code>' +
        '</div>' +
        '</div>'
    document.body.append(box)
}

