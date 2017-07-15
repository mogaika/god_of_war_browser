package web

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strconv"

	"github.com/gorilla/mux"

	file_vpk "github.com/mogaika/god_of_war_browser/pack/vpk"
	file_wad "github.com/mogaika/god_of_war_browser/pack/wad"
	file_vagp "github.com/mogaika/god_of_war_browser/ps2/vagp"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func HandlerAjaxPack(w http.ResponseWriter, r *http.Request) {
	files := ServerPack.GetFileNamesList()
	sort.Strings(files)
	webutils.WriteJson(w, files)
}

func HandlerAjaxPackFile(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	data, err := ServerPack.GetInstance(file)
	if err != nil {
		log.Printf("Error getting file from pack: %v", err)
		webutils.WriteError(w, err)
	} else {
		webutils.WriteJson(w, data)
	}
}

func HandlerAjaxPackFileParam(w http.ResponseWriter, r *http.Request) {
	type Result struct {
		Node *file_wad.WadNode
		Data interface{}
	}

	file := mux.Vars(r)["file"]
	param := mux.Vars(r)["param"]
	data, err := ServerPack.GetInstance(file)
	if err != nil {
		log.Printf("Error getting file from pack: %v", err)
		webutils.WriteError(w, err)
	} else {
		switch file[len(file)-4:] {
		case ".WAD":
			wad := data.(*file_wad.Wad)
			id, err := strconv.Atoi(param)
			if err != nil {
				webutils.WriteError(w, fmt.Errorf("param '%s' is not integer", param))
			} else {
				node := wad.Node(id).ResolveLink()
				if node != nil {
					data, err := wad.Get(node.Id)
					if err == nil {
						val, err := data.Marshal(wad, node)
						if err != nil {
							webutils.WriteError(w, fmt.Errorf("Error Marshaling node %d from %s: %v", id, file, err.(error)))
						} else {
							webutils.WriteJson(w, &Result{Node: node, Data: val})
						}
					} else {
						webutils.WriteError(w, fmt.Errorf("File %s-%d[%s] reading error: %v", file, id, wad.Nodes[id].Name, err))
					}
				} else {
					webutils.WriteError(w, fmt.Errorf("Cannot find node %d in %s", id, wad.Name))
				}
			}
		default:
			webutils.WriteError(w, fmt.Errorf("File %s not contain subdata", file))
		}
	}
}

func HandlerDumpPackFile(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	_, reader, err := ServerPack.GetFileReader(file)
	if err == nil {
		webutils.WriteFile(w, reader, file)
	} else {
		fmt.Fprintf(w, "Error getting file reader: %v", err)
	}
}

func HandlerDumpPackParamFile(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	param := mux.Vars(r)["param"]
	data, err := ServerPack.GetInstance(file)
	if err != nil {
		log.Printf("Error getting file from pack: %v", err)
		webutils.WriteError(w, err)
	} else {
		switch file[len(file)-4:] {
		case ".WAD":
			wad := data.(*file_wad.Wad)
			id, err := strconv.Atoi(param)
			if err != nil {
				webutils.WriteError(w, fmt.Errorf("param '%s' is not integer", param))
			} else {
				if rdr, err := wad.GetFileReader(id); err == nil {
					webutils.WriteFile(w, rdr, wad.Nodes[id].Name)
				} else {
					webutils.WriteError(w, fmt.Errorf("cannot get wad '%s' file %d reader", file, id))
				}
			}
		case ".VAG":
			if wav, err := data.(*file_vagp.VAGP).AsWave(); err != nil {
				webutils.WriteError(w, fmt.Errorf("Error converting to wav: %v", err))
			} else {
				webutils.WriteFile(w, wav, file+".WAV")
			}
		case ".VPK":
			vpk := data.(*file_vpk.VPK)
			_, fr, err := ServerPack.GetFileReader(file)
			if err != nil {
				panic(err)
			}
			var buf bytes.Buffer
			_, err = vpk.AsWave(fr, &buf)
			if err != nil {
				webutils.WriteError(w, fmt.Errorf("Error converting to wav: %v", err))
			} else {
				webutils.WriteFile(w, &buf, file+".WAV")
			}
		default:
			webutils.WriteError(w, fmt.Errorf("File %s not contain subdata", file))
		}
	}
}

func HandlerDumpPackParamSubFile(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	param := mux.Vars(r)["param"]
	subfile := mux.Vars(r)["subfile"]
	data, err := ServerPack.GetInstance(file)
	if err != nil {
		log.Printf("Error getting file from pack: %v", err)
		webutils.WriteError(w, err)
	} else {
		switch file[len(file)-4:] {
		case ".WAD":
			wad := data.(*file_wad.Wad)
			id, err := strconv.Atoi(param)
			if err != nil {
				webutils.WriteError(w, fmt.Errorf("param '%s' is not integer", param))
			} else {
				node := wad.Node(id).ResolveLink()
				if node != nil {
					data, err := wad.Get(node.Id)
					if err == nil {
						rt := reflect.TypeOf(data)
						method, has := rt.MethodByName("SubfileGetter")
						if !has {
							webutils.WriteError(w, fmt.Errorf("Error: %s has not func SubfileGetter", rt.Name()))
						} else {
							method.Func.Call([]reflect.Value{
								reflect.ValueOf(data),
								reflect.ValueOf(w),
								reflect.ValueOf(r),
								reflect.ValueOf(wad),
								reflect.ValueOf(node),
								reflect.ValueOf(subfile),
							}[:])
						}
					} else {
						webutils.WriteError(w, fmt.Errorf("File %s-%d[%s] reading error: %v", file, id, wad.Nodes[id].Name, err))
					}
				} else {
					webutils.WriteError(w, fmt.Errorf("Cannot find node %d in %s", id, wad.Name))
				}
			}
		default:
			webutils.WriteError(w, fmt.Errorf("File %s not contain subdata", file))
		}
	}
}
