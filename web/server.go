package web

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/mogaika/god_of_war_browser/vfs"
)

var ServerDirectory vfs.Directory
var DriverDirectory vfs.Directory
var wsUpgrader = websocket.Upgrader{}

func StartServer(addr string, packsDir vfs.Directory, driver vfs.Directory, webPath string) error {
	ServerDirectory = packsDir
	DriverDirectory = driver

	r := mux.NewRouter()
	r.HandleFunc("/archive/get/{name}", HandleArchive)
	r.HandleFunc("/action/{file}/{param}/{action}", HandlerActionPackFileParam)
	r.HandleFunc("/json/pack/{file}/{param}", HandlerAjaxPackFileParam)
	r.HandleFunc("/json/pack/{file}", HandlerAjaxPackFile)
	r.HandleFunc("/json/pack", HandlerAjaxPack)
	r.HandleFunc("/json/fs", HandlerAjaxFs)
	r.HandleFunc("/dump/pack/{file}/{param}", HandlerDumpPackParamFile)
	r.HandleFunc("/dump/pack/{file}", HandlerDumpPackFile)
	r.HandleFunc("/dump/fs/{file}", HandlerDumpFsFile)
	r.HandleFunc("/delete/pack/{file}", HandlerDeletePackFile)
	r.HandleFunc("/upload/pack/{file}", HandlerUploadPackFile)
	r.HandleFunc("/upload/pack/{file}/{param}", HandlerUploadPackFileParam)
	r.HandleFunc("/ws/status", HandlerWebsocketStatus)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(path.Join(webPath, "data"))))

	h := handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(r)
	h = handlers.LoggingHandler(os.Stdout, h)

	log.Printf("[web] Starting server %v", addr)

	return http.ListenAndServe(addr, h)
}
