package e7s

import (
	"sync"
)

type disposeFunc func(c *Context)

type middle func(c *Context)

type disposeRouters map[string]disposeFunc

type Router struct {
	router          disposeRouters
	middle          []middle
	handlersRWMutex sync.RWMutex
}

var routers *Router

func NewRouter() *Router {
	routers = &Router{
		router: make(disposeRouters),
		middle: make([]middle, 1),
	}
	return routers
}

func (r *Router) Register(key string, value disposeFunc) *Router {
	r.handlersRWMutex.Lock()
	defer r.handlersRWMutex.Unlock()
	r.router[key] = value
	routers = r
	return r
}

func (r *Router) getHandlers(key string) (value disposeFunc, ok bool) {
	if len(r.router) <= 0 {
		return nil, false
	}
	r.handlersRWMutex.RLock()
	defer r.handlersRWMutex.RUnlock()
	value, ok = r.router[key]
	return
}

func (r *Router) Use(middle middle) {
	r.middle = append(r.middle, middle)
	routers = r
}
