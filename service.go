package e7s

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/silenceper/log"
	"time"
)

const (
	// 用户连接超时时间
	HeartbeatExpirationTime = 6 * 60
)

// Client 客户端实例
type Client struct {
	Addr          string          // 客户端地址
	Socket        *websocket.Conn // 用户连接
	Clients       string          // 客户端标识
	Send          chan []byte     // 待发送的数据
	UserId        string          // 用户Id，用户登录以后才有
	FirstTime     uint64          // 首次连接时间
	HeartbeatTime uint64          // 用户上次心跳时间
	LoginTime     uint64          // 登录时间 登录以后才有
	Token         string          //登陆token
}

//消息体
type Msg struct {
	Api     string
	C       string
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
			Response(c, SERVER_ERROR, nil)
			return
		}
	}()
	var message Msg
	err := json.Unmarshal(msg, &message)
	if err != nil {
		Response(c, REQUEST_PARAMRTER_ERROR, nil)
		return
	}
	if message.Api == "" || message.C == "" {
		Response(c, REQUEST_PARAMRTER_ERROR, nil)
		return
	}
	controllers := message.Api + "_" + message.C
	context := &E7sContext{
		Client:  c,
		Manager: Managers,
		Request: message.Request,
		Api:     message.Api,
		C:       message.C,
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
		Response(c, ROUTE_EROOR, nil)
		return
	}
}
