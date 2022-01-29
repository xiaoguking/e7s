package e7s

import (
	"github.com/silenceper/log"
	"net/http"
)

type e7s struct {
	Router *Router
}

func NewE7s() *e7s {
	return &e7s{
		Router: NewRouter(),
	}
}

func (e *e7s) Run(port string) error {

	http.HandleFunc("/", Handle)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
