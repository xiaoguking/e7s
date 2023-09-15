package e7s

import (
	"github.com/silenceper/log"
	"net/http"
)

type E7s struct {
	Router *Router
	Root   string
}

func NewE7s(root string) *E7s {
	return &E7s{
		Router: NewRouter(),
		Root:   root,
	}
}

func (e *E7s) Run(port string) error {
	go managers.start()
	http.HandleFunc(e.Root, handle)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
