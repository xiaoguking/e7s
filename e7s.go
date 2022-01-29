package e7s

import (
	"github.com/silenceper/log"
	"net/http"
)

type E7s struct {
	Router *Router
}

func NewE7s() *E7s {
	return &E7s{
		Router: NewRouter(),
	}
}

func (e *E7s) Run(port string) error {

	http.HandleFunc("/", Handle)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
