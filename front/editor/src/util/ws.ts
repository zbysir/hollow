export default class Ws {
    private callback: ((bs: string) => void)[]
    conn: WebSocket

    constructor(key: string) {
        this.callback = []

        this.conn = new WebSocket("ws://localhost:9091/ws/"+key);
        this.conn.onclose = function (evt) {
            // var item = document.createElement("div");
            // item.innerHTML = "<b>Connection closed.</b>";
            console.log('close');
        };
        this.conn.onmessage = (evt) => {
            let messages = evt.data.split('\n');
            for (let i = 0; i < messages.length; i++) {
                // var item = document.createElement("div");
                // item.innerText = messages[i];
                this.callback.forEach(c => {
                    c(messages[i])
                })
            }
        };
    }

    public Register(f: (bs: string) => void) {
        this.callback.push(f)
    }


    public Row(): WebSocket {
        return this.conn
    }


}
