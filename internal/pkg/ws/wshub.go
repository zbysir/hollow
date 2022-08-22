package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"sync"
	"time"
)

type WsHub struct {
	conns map[string]*websocket.Conn
	l     sync.Mutex
}

func NewHub() *WsHub {
	return &WsHub{
		conns: map[string]*websocket.Conn{},
		l:     sync.Mutex{},
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
func (h *WsHub) Add(key string, conn *websocket.Conn) {
	if key == "" {
		key = fmt.Sprintf("%v", rand.Int63())
	}
	if o, ok := h.conns[key]; ok {
		_ = o.Close()
	}

	h.conns[key] = conn
}

func (h *WsHub) Send(key string, body []byte) error {
	if o, ok := h.conns[key]; ok {
		err := o.WriteMessage(1, body)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *WsHub) SendAll(body []byte) error {
	for _, o := range h.conns {
		err := o.WriteMessage(1, body)
		if err != nil {
			return err
		}
	}

	return nil
}
