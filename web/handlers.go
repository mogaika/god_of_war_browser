package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"

	file_wad "github.com/mogaika/god_of_war_browser/pack/wad"
)

func writeFile(w http.ResponseWriter, in io.Reader, name string) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
	io.Copy(w, in)
}

func writeJson(w http.ResponseWriter, data interface{}) {
	res, err := json.Marshal(data)
	if err != nil {
		writeError(w, err)
	} else {
		writeResult(w, res)
	}
}

func writeResult(w http.ResponseWriter, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		log.Printf("Error when writing reponse: %v", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	type jError struct {
		Error string `json:"error"`
	}
	data, err := json.Marshal(&jError{Error: err.Error()})
	if err == nil {
		log.Printf("HERR: %v", string(data))
		writeResult(w, data)
	} else {
		log.Printf("Error marshaling error '%v': %v", err, data)
	}
}

func HandlerAjaxPack(w http.ResponseWriter, r *http.Request) {
	writeJson(w, ServerPack)
}

func HandlerAjaxPackFile(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	data, err := ServerPack.Get(file)
	if err != nil {
		log.Printf("Error getting file from pack: %v", err)
		writeError(w, err)
	} else {
		writeJson(w, data)
	}
}

func HandlerAjaxPackFileParam(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	param := mux.Vars(r)["param"]
	data, err := ServerPack.Get(file)
	if err != nil {
		log.Printf("Error getting file from pack: %v", err)
		writeError(w, err)
	} else {
		switch file[len(file)-4:] {
		case ".WAD":
			wad := data.(*file_wad.Wad)
			id, err := strconv.Atoi(param)
			if err != nil {
				writeError(w, fmt.Errorf("param '%s' is not integer", param))
			} else {
				data, err := wad.Get(id)
				if err == nil {
					val := reflect.ValueOf(data)
					ajaxmarshal := val.MethodByName("AjaxMarshal")
					if ajaxmarshal.IsValid() {
						retval := ajaxmarshal.Call([]reflect.Value{
							reflect.ValueOf(wad),
							reflect.ValueOf(wad.Nodes[id]),
						})

						val := retval[0].Interface()
						err := retval[1].Interface()
						if err != nil {
							writeError(w, fmt.Errorf("Error AjaxMarshaling node %d from %s: %v", id, file, err.(error)))
						} else {
							writeResult(w, val.([]byte))
						}
					} else {
						writeJson(w, data)
					}
				} else {
					writeError(w, fmt.Errorf("File %s->%d[%s] reading error: %v", file, id, wad.Nodes[id].Name, err))
				}
			}
		default:
			writeError(w, fmt.Errorf("File %s not contain subdata", file))
		}
	}
}

func HandlerDumpPackFile(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	reader, err := ServerPack.GetFileReader(file)
	if err == nil {
		writeFile(w, reader, file)
	} else {
		fmt.Fprintf(w, "Error getting file reader: %v", err)
	}
}

func HandlerDumpPackParamFile(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	param := mux.Vars(r)["param"]
	data, err := ServerPack.Get(file)
	if err != nil {
		log.Printf("Error getting file from pack: %v", err)
		writeError(w, err)
	} else {
		switch file[len(file)-4:] {
		case ".WAD":
			wad := data.(*file_wad.Wad)
			id, err := strconv.Atoi(param)
			if err != nil {
				writeError(w, fmt.Errorf("param '%s' is not integer", param))
			} else {
				if rdr, err := wad.GetFileReader(id); err == nil {
					writeFile(w, rdr, wad.Nodes[id].Name)
				} else {
					writeError(w, fmt.Errorf("cannot get wad '%s' file %d reader", file, id))
				}
			}
		default:
			writeError(w, fmt.Errorf("File %s not contain subdata", file))
		}
	}
}
