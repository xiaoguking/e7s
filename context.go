package e7s

import (
	"encoding/json"
	"github.com/silenceper/log"
	"sync"
)

const (
	ROUTE_EROOR             = -1001
	REQUEST_PARAMRTER_ERROR = -1002
	SERVER_ERROR            = -1003
)

type Context struct {
	//websocket client
	Client *Client
	//ClientManager
	Manager *ClientManager
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
	Cmd      string      `json:"cmd,omitempty"`
	Status   int         `json:"status"`
	Response interface{} `json:"response,omitempty"`
}

func (c *Context) GetRequestUid() (string, bool) {
	uid := c.Client.UserId
	if uid == "" {
		return "", false
	}
	return uid, true
}

func (c *Context) JSON(status int, obj interface{}) {
	res := response{}
	res.Cmd = c.Api + "_" + c.C
	res.Status = status
	res.Response = obj
	data, err := json.Marshal(res)
	if err != nil {
		log.Error(err.Error())
	}
	c.Client.Send <- data
}

func Response(c *Client, status int, obj interface{}) {
	res := response{}
	res.Status = status
	res.Response = obj
	data, err := json.Marshal(res)
	if err != nil {
		log.Error(err.Error())
	}
	c.Send <- data
}

func (c *Context) GetRequest(key string, defaults string) interface{} {
	c.cLock.RLock()
	defer c.cLock.RUnlock()
	if val, ok := c.Request[key]; ok == false {
		return defaults
	} else {
		return val
	}
}
