package e7s

import (
	"github.com/fvbock/endless"
	"github.com/gorilla/mux"
	"github.com/silenceper/log"
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
	mux1 := mux.NewRouter()
	mux1.HandleFunc(e.Root, handle)
	//http.HandleFunc(e.Root, handle)
	if err := endless.ListenAndServe(":"+port, mux1); err != nil {
		log.Error(err.Error())
		return err

	}
	//if err := http.ListenAndServe(":"+port, nil); err != nil {
	//	log.Error(err.Error())
	//	return err
	//}
	return nil
}
