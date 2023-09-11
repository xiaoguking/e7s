package e7s

import (
	"github.com/fvbock/endless"
	"github.com/gorilla/mux"
)

type E7s struct {
	Router *Router
	Root   string
}

func NewE7s() *E7s {
	return &E7s{
		Router: NewRouter(),
		Root:   "/",
	}
}

func (e *E7s) Run(port string) error {
	go managers.start()
	mux1 := mux.NewRouter()
	mux1.HandleFunc(e.Root, handle)

	s := endless.NewServer(":"+port, mux1)
	err := s.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}
