package e7s

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/silenceper/log"
	"runtime/debug"
)

const (
	// 用户连接超时时间
	HeartbeatExpirationTime = 6 * 60
)

// 客户端实例
type Client struct {
	Addr          string          // 客户端地址
	Socket        *websocket.Conn // 用户连接
	Clients       string          //客户端标识
	Send          chan []byte     // 待发送的数据
	UserId        string          // 用户Id，用户登录以后才有
	FirstTime     uint64          // 首次连接时间
	HeartbeatTime uint64          // 用户上次心跳时间
	LoginTime     uint64          // 登录时间 登录以后才有
}

//消息体
type Msg struct {
	Cmd     string
	Request map[string]interface{}
}

// 初始化
func NewClient(addr string, socket *websocket.Conn, firstTime uint64, clients string) (client *Client) {

	client = &Client{
		Addr:          addr,
		Socket:        socket,
		Clients:       clients,
		Send:          make(chan []byte, 100),
		FirstTime:     firstTime,
		HeartbeatTime: firstTime,
	}
	return
}

//写消息
func (c *Client) writer() {
	for message := range c.Send {
		c.Socket.WriteMessage(websocket.TextMessage, message)
	}
	c.Socket.Close()
}

//读取消息
func (c *Client) reader() {
	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			break
		}
		log.Info("Description The client message was received " + string(message))
		onmessage(message, c)
	}
}

func onmessage(msg []byte, c *Client) {
	defer func() {
		if err := recover(); err != nil {
			log.Debug(string(debug.Stack()))
			c.Send <- []byte("server error")
			return
		}
	}()
	var message Msg
	err := json.Unmarshal(msg, &message)
	if err != nil {
		c.Send <- []byte("params error")
		return
	}
	controllers := message.Cmd
	context := &E7sContext{
		Client:  c,
		Manager: managers,
		Request: message.Request,
		Cmd:     controllers,
		Next:    true,
	}

	if value, ok := routers.getHandlers(controllers); ok {
		if len(routers.middle) > 0 {
			for _, v := range routers.middle {
				if context.Next == true && v != nil {
					v(context)
				}
			}
		}
		if context.Next {
			value(context)
		}
	} else {
		log.Debug("websocket router not")
		c.Send <- []byte("websocket router not")
		return
	}
}
