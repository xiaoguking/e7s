package e7s

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"time"
)

const (
	// HeartbeatExpirationTime 用户连接超时时间
	HeartbeatExpirationTime = 6 * 60
)

// Client 客户端实例
type client struct {
	addr          string          // 客户端地址
	socket        *websocket.Conn // 用户连接
	clients       string          // 客户端标识
	send          chan []byte     // 待发送的数据
	userId        int             // 用户Id，用户登录以后才有
	firstTime     uint64          // 首次连接时间
	heartbeatTime uint64          // 用户上次心跳时间
	loginTime     uint64          // 登录时间 登录以后才有
	token         string          //登陆token
}

// Request 消息体
type request struct {
	Api     string
	C       string
	Request map[string]interface{}
}

// NewClient 初始化
func newClient(addr string, socket *websocket.Conn, firstTime uint64, clients string) *client {

	return &client{
		addr:          addr,
		socket:        socket,
		clients:       clients,
		send:          make(chan []byte, 100),
		firstTime:     firstTime,
		heartbeatTime: firstTime,
	}
}

//写消息
func (c *client) writer() {
	for message := range c.send {
		c.socket.WriteMessage(websocket.TextMessage, message)
	}
	c.socket.Close()
}

//读取消息
func (c *client) reader() {
	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			break
		}
		onmessage(message, c)
	}
}

func onmessage(msg []byte, c *client) {
	defer func() {
		if err := recover(); err != nil {
			sendResponse(c, ServerError, nil)
			return
		}
	}()
	var message request
	err := json.Unmarshal(msg, &message)
	if err != nil {
		sendResponse(c, RequestParamsError, nil)
		return
	}
	if message.Api == "" || message.C == "" {
		sendResponse(c, RequestParamsError, nil)
		return
	}
	controllers := message.Api + "_" + message.C
	context := &Context{
		client:  c,
		manager: managers,
		Request: message.Request,
		Api:     message.Api,
		C:       message.C,
		Next:    true,
	}

	if value, ok := routers.getHandlers(controllers); ok {
		c.heartbeatTime = uint64(time.Now().Unix())
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
		sendResponse(c, RouteError, nil)
		return
	}
}
