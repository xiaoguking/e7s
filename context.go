package e7s

import (
	"encoding/json"
	"github.com/silenceper/log"
	"sync"
	"time"
	//"fmt"
)

const (
	RouteError         = -1001
	RequestParamsError = -1002
	ServerError        = -1003
)

type Context struct {
	//websocket client
	client *Client
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
	Api    string      `json:"api,omitempty"`
	Status int         `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}

func sendResponse(c *Client, status int, obj interface{}) {
	res := response{}
	res.Status = status
	res.Data = obj
	data, err := json.Marshal(res)
	if err != nil {
		log.Error(err.Error())
	}
	c.send <- data
}

// GetUid 获取uid
func (c *Context) GetUid() string {
	return c.client.UserId
}

// IsLogin 是否登录
func (c *Context) IsLogin() bool {
	if c.client.UserId == "" {
		return false
	} else {
		return true
	}
}

//JSON 返回JSON 数据
func (c *Context) JSON(status int, obj interface{}) {
	res := response{}
	res.Api = c.Api + "_" + c.C
	res.Status = status
	res.Data = obj
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

// GetQuery 查询请求参数
func (c *Context) GetQuery(key string) string {
	c.cLock.RLock()
	defer c.cLock.RUnlock()
	return StructToURLValues(c.Request, key)
}

// Login 登录事件
func (c *Context) Login(uid string, time int) {
	c.client.UserId = uid
	c.client.LoginTime = uint64(time)
	uidClient := &Login{
		uid: uid,
		c:   c.client,
	}
	managers.login <- uidClient
}

// Logout 退出事件
func (c *Context) Logout() {
	managers.loginOut <- c.client
}

// BanUid 封号
func (c *Context) BanUid(uid string) {
	managers.uidBan <- uid
}

// SendToUid 单独向uid发送消息
func SendToUid(uid string, msg []byte) {
	managers.clientsLock.RLock()
	defer managers.clientsLock.RUnlock()
	uidClient := managers.getUserClient(uid)
	if uidClient != nil {
		uidClient.send <- msg
	}
}

// SendToUids 向多个uid发送消息
func SendToUids(uid []string, msg []byte) {
	managers.clientsLock.RLock()
	defer managers.clientsLock.RUnlock()
	for _, v := range uid {
		uidClient := managers.getUserClient(v)
		if uidClient != nil {
			uidClient.send <- msg
		}
	}
}

// SendOther 向其他全部成员发送数据
func SendOther(message []byte) {
}

// SendAll 发送广播
func SendAll(form string, message []byte) {
	messages := &broadcastMessage{}
	messages.From = form
	messages.Message = message
	managers.broadcast <- messages
}

// GetOnlineLen 获取当前在线的人数
func GetOnlineLen() int {
	managers.clientsLock.RLock()
	defer managers.clientsLock.RUnlock()
	return len(managers.clients)
}

// HeartbeatCheck 心跳检测
func HeartbeatCheck(heartbeatTime uint64) {
	managers.clientsLock.RLock()
	defer managers.clientsLock.RUnlock()
	for k := range managers.clients {
		if k != nil && uint64(time.Now().Unix())-k.HeartbeatTime > heartbeatTime {
			//managers.loginOut <- k
			managers.unregister <- k
			k.Socket.Close()
		}
	}
}
