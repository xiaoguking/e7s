package e7s

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var ws = &websocket.Upgrader{ReadBufferSize: 512, WriteBufferSize: 512, CheckOrigin: func(r *http.Request) bool { return true }}

var managers = newClientManager()

func handle(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Server", " Server/1.0")
	conn, err := ws.Upgrade(w, r, w.Header())
	if err != nil {
		return
	}
	addr := conn.RemoteAddr().String()
	currentTime := uint64(time.Now().Unix())
	clients := uniqueId()
	c := newClient(addr, conn, currentTime, clients)
	managers.register <- c
	go c.writer()
	go c.reader()

	defer func() {
		managers.unregister <- c
		c.socket.Close()
	}()
}
