package e7s

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

const (
	// HeartbeatExpirationTime 用户连接超时时间
	HeartbeatExpirationTime = 6 * 60
)

// Client 客户端实例
type Client struct {
	Addr          string          // 客户端地址
	Socket        *websocket.Conn // 用户连接
	Clients       string          // 客户端标识
	send          chan []byte     // 待发送的数据
	UserId        string          // 用户Id，用户登录以后才有
	FirstTime     uint64          // 首次连接时间
	HeartbeatTime uint64          // 用户上次心跳时间
	LoginTime     uint64          // 登录时间 登录以后才有
	Token         string          //登陆token
}

// NewClient 初始化
func newClient(addr string, socket *websocket.Conn, firstTime uint64, Clients string) *Client {

	return &Client{
		Addr:          addr,
		Socket:        socket,
		Clients:       Clients,
		send:          make(chan []byte, 100),
		FirstTime:     firstTime,
		HeartbeatTime: firstTime,
	}
}

//写消息
func (c *Client) writer() {
	for message := range c.send {
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
		onmessage(message, c)
	}
}

func onmessage(msg []byte, c *Client) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			return
		}
	}()
	message := make(map[string]interface{})
	err := json.Unmarshal(msg, &message)
	if err != nil {
		sendResponse(c, RequestParamsError, nil)
		return
	}
	if _, ok := message["api"]; !ok {
		sendResponse(c, RequestParamsError, nil)
		return
	}
	if _, ok := message["c"]; !ok {
		sendResponse(c, RequestParamsError, nil)
		return
	}
	api := StructToURLValues(message, "api")
	cc := StructToURLValues(message, "c")
	controllers := api + "_" + cc
	context := &Context{
		client:  c,
		Request: message,
		Api:     api,
		C:       cc,
		Next:    true,
	}

	if value, ok := routers.getHandlers(controllers); ok {
		c.HeartbeatTime = uint64(time.Now().Unix())
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
