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
	file := mux.Vars(r)["file"]
	param := mux.Vars(r)["param"]
	data, err := ServerPack.GetInstance(file)
	if err != nil {
		log.Printf("Error getting file from pack: %v", err)
		webutils.WriteError(w, err)
	} else {
		switch data.(type) {
		case *file_wad.Wad:
			wad := data.(*file_wad.Wad)
			id, err := strconv.Atoi(param)
			if err != nil {
				webutils.WriteError(w, fmt.Errorf("param '%s' is not integer", param))
			} else {
				if err := wad.WebHandlerForNodeByTagId(w, file_wad.TagId(id)); err != nil {
					webutils.WriteError(w, fmt.Errorf("wad web handler return error: %v", err))
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
		switch data.(type) {
		case *file_wad.Wad:
			wad := data.(*file_wad.Wad)
			id, err := strconv.Atoi(param)
			if err != nil {
				webutils.WriteError(w, fmt.Errorf("param '%s' is not integer", param))
			} else {
				tag := wad.GetTagById(file_wad.TagId(id))
				webutils.WriteFile(w, bytes.NewBuffer(tag.Data), tag.Name)
			}
		case *file_vagp.VAGP:
			if wav, err := data.(*file_vagp.VAGP).AsWave(); err != nil {
				webutils.WriteError(w, fmt.Errorf("Error converting to wav: %v", err))
			} else {
				webutils.WriteFile(w, wav, file+".WAV")
			}
		case *file_vpk.VPK:
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

func HandlerAactionPackFileParam(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	param := mux.Vars(r)["param"]
	action := mux.Vars(r)["action"]
	data, err := ServerPack.GetInstance(file)
	if err != nil {
		log.Printf("Error getting file from pack: %v", err)
		webutils.WriteError(w, err)
	} else {
		switch data.(type) {
		case *file_wad.Wad:
			wad := data.(*file_wad.Wad)
			id, err := strconv.Atoi(param)
			if err != nil {
				webutils.WriteError(w, fmt.Errorf("param '%s' is not integer", param))
			} else {
				id := file_wad.TagId(id)
				if inst, _, err := wad.GetInstanceFromTag(id); err == nil {
					rt := reflect.TypeOf(inst)
					method, has := rt.MethodByName("HttpAction")
					if !has {
						webutils.WriteError(w, fmt.Errorf("Error: %s has not func SubfileGetter", rt.Name()))
					} else {
						method.Func.Call([]reflect.Value{
							reflect.ValueOf(inst),
							reflect.ValueOf(wad.GetNodeResourceByTagId(id)),
							reflect.ValueOf(w),
							reflect.ValueOf(r),
							reflect.ValueOf(action),
						}[:])
					}
				} else {
					webutils.WriteError(w, fmt.Errorf("File %s-%d instance getting error: %v", file, id, err))
				}
			}
		default:
			webutils.WriteError(w, fmt.Errorf("File %s not contain subdata", file))
		}
	}
}
