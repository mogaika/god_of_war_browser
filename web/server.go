package web

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/mogaika/god_of_war_browser/pack"
)

var ServerPack *pack.Pack

func StartServer(addr string, _pack *pack.Pack) error {
	ServerPack = _pack

	r := mux.NewRouter()
	r.HandleFunc("/json/pack/{file}/{param}", HandlerAjaxPackFileParam)
	r.HandleFunc("/json/pack/{file}", HandlerAjaxPackFile)
	r.HandleFunc("/json/pack", HandlerAjaxPack)

	r.HandleFunc("/dump/pack/{file}/{param}/{subfile}", HandlerDumpPackParamSubFile)
	r.HandleFunc("/dump/pack/{file}/{param}", HandlerDumpPackParamFile)
	r.HandleFunc("/dump/pack/{file}", HandlerDumpPackFile)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/data")))

	h := handlers.RecoveryHandler()(r)
	h = handlers.LoggingHandler(os.Stdout, r)

	log.Printf("Starting server %v", addr)

	return http.ListenAndServe(addr, h)
}
