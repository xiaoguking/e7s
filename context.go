package e7s

import (
	"encoding/json"
	"github.com/silenceper/log"
	"sync"
)

type E7sContext struct {
	//websocket client
	Client *Client
	//ClientManager
	Manager *ClientManager
	//websocket request
	Request map[string]interface{}
	//router
	Cmd string
	//请求继续
	Next bool
	//错误日志
	Error string
	//读写锁
	cLock sync.RWMutex
}

type response struct {
	Cmd      string      `json:"cmd"`
	Status   int         `json:"status"`
	Response interface{} `json:"response,omitempty"`
}

//向客户端 发送JSON消息
func (c *E7sContext) JSON(status int, obj interface{}) {
	res := response{}
	res.Cmd = c.Cmd
	res.Status = status
	res.Response = obj
	data, err := json.Marshal(res)
	if err != nil {
		log.Error(err.Error())
	}
	c.Client.Send <- data
}

//获取请求参数
func (c *E7sContext) GetRequest(key string) interface{} {
	c.cLock.RLock()
	defer c.cLock.RUnlock()
	return c.Request[key]
}
