package e7s

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var wu = &websocket.Upgrader{ReadBufferSize: 512, WriteBufferSize: 512, CheckOrigin: func(r *http.Request) bool { return true }}

var managers = NewClientManager()

func Handle(w http.ResponseWriter, r *http.Request) {

	go managers.Start()

	w.Header().Set("Server", " Server/1.0")
	ws, err := wu.Upgrade(w, r, w.Header())
	if err != nil {
		return
	}
	addr := ws.RemoteAddr().String()
	currentTime := uint64(time.Now().Unix())
	clients := uniqueId()
	c := NewClient(addr, ws, currentTime, clients)
	managers.register <- c
	go c.writer()
	c.reader()
	defer func() {
		managers.unregister <- c
		c.socket.Close()
	}()
}
