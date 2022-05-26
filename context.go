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

func (c *E7sContext) GetUid() (string, bool) {
	uid := c.Client.UserId
	if uid == "" {
		return "", false
	}
	return uid, true
}

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

func (c *E7sContext) GetRequest(key string, defaults string) interface{} {
	c.cLock.RLock()
	defer c.cLock.RUnlock()
	if val, ok := c.Request[key]; ok == false {
		return defaults
	} else {
		return val
	}
}
