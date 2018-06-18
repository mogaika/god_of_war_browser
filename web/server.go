package web

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/mogaika/god_of_war_browser/vfs"
)

var ServerDirectory vfs.Directory

func StartServer(addr string, d vfs.Directory, webPath string) error {
	ServerDirectory = d

	r := mux.NewRouter()
	r.HandleFunc("/action/{file}/{param}/{action}", HandlerActionPackFileParam)
	r.HandleFunc("/json/pack/{file}/{param}", HandlerAjaxPackFileParam)
	r.HandleFunc("/json/pack/{file}", HandlerAjaxPackFile)
	r.HandleFunc("/json/pack", HandlerAjaxPack)
	r.HandleFunc("/dump/pack/{file}/{param}", HandlerDumpPackParamFile)
	r.HandleFunc("/dump/pack/{file}", HandlerDumpPackFile)
	r.HandleFunc("/upload/pack/{file}", HandlerUploadPackFile)
	r.HandleFunc("/upload/pack/{file}/{param}", HandlerUploadPackFileParam)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(path.Join(webPath, "data"))))

	h := handlers.RecoveryHandler()(r)
	h = handlers.LoggingHandler(os.Stdout, r)

	log.Printf("[web] Starting server %v", addr)

	return http.ListenAndServe(addr, h)
}
