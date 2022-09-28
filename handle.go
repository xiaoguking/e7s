package e7s

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var wu = &websocket.Upgrader{ReadBufferSize: 512, WriteBufferSize: 512, CheckOrigin: func(r *http.Request) bool { return true }}

var Managers = NewClientManager()

func Handle(w http.ResponseWriter, r *http.Request) {

	go Managers.Start()

	w.Header().Set("Server", " Server/1.0")
	ws, err := wu.Upgrade(w, r, w.Header())
	if err != nil {
		return
	}
	//log.Info("webSocket The client establishes a connection:" + ws.RemoteAddr().String())

	addr := ws.RemoteAddr().String()
	currentTime := uint64(time.Now().Unix())
	clients := uniqueId()
	c := NewClient(addr, ws, currentTime, clients)
	Managers.Register <- c
	go c.writer()
	c.reader()
	defer func() {
		//log.Info("websocket The client is disconnected. Procedure" + ws.RemoteAddr().String())
		Managers.Unregister <- c
		c.Socket.Close()
	}()
}
