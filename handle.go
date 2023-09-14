package e7s

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var ws = &websocket.Upgrader{ReadBufferSize: 512, WriteBufferSize: 512, CheckOrigin: func(r *http.Request) bool { return true }}

var managers = newClientManager()

func handle(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			return
		}
	}()
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
	go c.reader()
	c.writer()

	defer func() {
		managers.unregister <- c
	}()
}
