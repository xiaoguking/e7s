package e7s

import (
	"github.com/gorilla/websocket"
	"github.com/silenceper/log"
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
	log.Info("webSocket 客户端建立连接:" + ws.RemoteAddr().String())

	addr := ws.RemoteAddr().String()
	currentTime := uint64(time.Now().Unix())
	c := NewClient(addr, ws, currentTime)
	managers.Register <- c
	go c.writer()
	c.reader()
	defer func() {
		log.Info("websocket 客户端断开连接" + ws.RemoteAddr().String())
		managers.Unregister <- c
		c.Socket.Close()
	}()
}
