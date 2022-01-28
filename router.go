package e7s

import (
	"sync"
)

type DisposeFunc func(c *E7sContext)

type Router struct {
	router          map[string]DisposeFunc
	handlersRWMutex sync.RWMutex
}

var routers *Router

func NewRouter() *Router {
	routers := &Router{
		router: make(map[string]DisposeFunc),
	}
	return routers
}

// 注册
func (r *Router) Register(key string, value DisposeFunc) *Router {
	r.handlersRWMutex.Lock()
	defer r.handlersRWMutex.Unlock()
	r.router[key] = value
	routers = r
	return r
}

//获取
func (r *Router) getHandlers(key string) (value DisposeFunc, ok bool) {
	if len(r.router) <= 0 {
		return nil, false
	}
	r.handlersRWMutex.RLock()
	defer r.handlersRWMutex.RUnlock()
	value, ok = r.router[key]
	return
}
