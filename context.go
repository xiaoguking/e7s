package e7s

import (
	"encoding/json"
	"github.com/silenceper/log"
	"sync"
)

const (
	RouteError         = -1001
	RequestParamsError = -1002
	ServerError        = -1003
)

type Context struct {
	//websocket client
	client *client
	//websocket request
	Request map[string]interface{}
	//router
	Api string
	//controller
	C string
	//请求继续
	Next bool
	//错误日志
	Error string
	//读写锁
	cLock sync.RWMutex
}

type response struct {
	Api      string      `json:"api,omitempty"`
	Status   int         `json:"status"`
	Response interface{} `json:"response,omitempty"`
}

func sendResponse(c *client, status int, obj interface{}) {
	res := response{}
	res.Status = status
	res.Response = obj
	data, err := json.Marshal(res)
	if err != nil {
		log.Error(err.Error())
	}
	c.send <- data
}

func (c *Context) GetRequestUid() int {
	return c.client.userId
}

func (c *Context) JSON(status int, obj interface{}) {
	res := response{}
	res.Api = c.Api + "_" + c.C
	res.Status = status
	res.Response = obj
	data, err := json.Marshal(res)
	if err != nil {
		log.Error(err.Error())
	}
	c.client.send <- data
}

func (c *Context) GetRequest(key string) interface{} {
	c.cLock.RLock()
	defer c.cLock.RUnlock()
	if val, ok := c.Request[key]; ok == false {
		return nil
	} else {
		return val
	}
}

// GetQuery 查询参数
func (c *Context) GetQuery(key string) string {
	c.cLock.RLock()
	defer c.cLock.RUnlock()
	return StructToURLValues(c.Request, key)
}

// Login 登录事件
func (c *Context) Login(uid int, time int) {
	c.client.userId = uid
	c.client.loginTime = uint64(time)
	uidClient := &Login{
		uid: uid,
		c:   c.client,
	}
	managers.login <- uidClient
}

func (c *Context) Logout() {
	managers.loginOut <- c.client
}

func (c *Context) BanUid(uid int) {
	managers.uidBan <- uid
}

// SendToUid 单独向uid发送消息
func (c *Context) SendToUid(uid int, msg []byte) {
	uidClient := managers.getUserClient(uid)
	uidClient.send <- msg
}

// SendToUids 向多个uid发送消息
func (c *Context) SendToUids(uid []int, msg []byte) {
	for _, v := range uid {
		uidClient := managers.getUserClient(v)
		uidClient.send <- msg
	}
}

// SendOther 向全部成员(除了自己)发送数据
func (c *Context) SendOther(message []byte) {

	clients := managers.getUserClients()
	for _, conn := range clients {
		if conn != c.client {
			conn.send <- message
		}
	}
}

// SendAll 发送广播
func (c *Context) SendAll(message []byte) {
	managers.broadcast <- message
}

// GetOnlineLen 获取当前在线的人数
func GetOnlineLen() (len int) {
	len = managers.getUOnlineLen()
	return
}
