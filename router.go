package e7s

import (
	"sync"
)

type DisposeFunc func(c *E7sContext)

type Middle func(c *E7sContext)

type DisposeRouters map[string]DisposeFunc

type Router struct {
	router          map[string]DisposeFunc
	middle          []Middle
	handlersRWMutex sync.RWMutex
}

var routers *Router

func NewRouter() *Router {
	routers = &Router{
		router: make(DisposeRouters),
		middle: make([]Middle, 0),
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

func (r *Router) Use(middle Middle) {
	r.middle = append(r.middle, middle)
	routers = r
}
