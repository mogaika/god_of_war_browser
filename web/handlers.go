package web

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/mogaika/god_of_war_browser/pack"
	file_vpk "github.com/mogaika/god_of_war_browser/pack/vpk"
	file_wad "github.com/mogaika/god_of_war_browser/pack/wad"
	file_vagp "github.com/mogaika/god_of_war_browser/ps2/vagp"
	"github.com/mogaika/god_of_war_browser/vfs"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func HandlerAjaxPack(w http.ResponseWriter, r *http.Request) {
	if files, err := ServerDirectory.List(); err != nil {
		webutils.WriteError(w, err)
	} else {
		sort.Strings(files)
		webutils.WriteJson(w, files)
	}
}

func HandlerAjaxPackFile(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	data, err := pack.GetInstanceHandler(ServerDirectory, file)
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
	data, err := pack.GetInstanceHandler(ServerDirectory, file)
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
	f, err := vfs.DirectoryGetFile(ServerDirectory, file)
	if err != nil {
		webutils.WriteError(w, err)
	}

	if reader, err := vfs.OpenFileAndGetReader(f, true); err == nil {
		webutils.WriteFile(w, reader, file)
		defer f.Close()
	} else {
		fmt.Fprintf(w, "Error getting file reader: %v", err)
	}
}

func HandlerDumpPackParamFile(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	param := mux.Vars(r)["param"]
	data, err := pack.GetInstanceHandler(ServerDirectory, file)
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
				wad.WebHandlerDumpTagData(w, file_wad.TagId(id))
			}
		case *file_vagp.VAGP:
			if wav, err := data.(*file_vagp.VAGP).AsWave(); err != nil {
				webutils.WriteError(w, fmt.Errorf("Error converting to wav: %v", err))
			} else {
				webutils.WriteFile(w, wav, file+".WAV")
			}
		case *file_vpk.VPK:
			vpk := data.(*file_vpk.VPK)

			f, err := vfs.DirectoryGetFile(ServerDirectory, file)
			if err != nil {
				webutils.WriteError(w, err)
			} else {
				fr, err := vfs.OpenFileAndGetReader(f, true)
				if err != nil {
					webutils.WriteError(w, err)
				}
				defer f.Close()

				var buf bytes.Buffer
				_, err = vpk.AsWave(fr, &buf)
				if err != nil {
					webutils.WriteError(w, fmt.Errorf("Error converting to wav: %v", err))
				} else {
					webutils.WriteFile(w, &buf, file+".WAV")
				}
			}
		default:
			webutils.WriteError(w, fmt.Errorf("File %s not contain subdata", file))
		}
	}
}

func HandlerActionPackFileParam(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	param := mux.Vars(r)["param"]
	action := mux.Vars(r)["action"]
	data, err := pack.GetInstanceHandler(ServerDirectory, file)
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
				if err := wad.WebHandlerCallResourceHttpAction(w, r, file_wad.TagId(id), action); err != nil {
					webutils.WriteError(w, fmt.Errorf("Wad handler error on %s-%d instance: %v", file, id, err))
				}
			}
		default:
			webutils.WriteError(w, fmt.Errorf("File %s not contain subdata", file))
		}
	}
}

func HandlerUploadPackFile(w http.ResponseWriter, r *http.Request) {
	targetFile := mux.Vars(r)["file"]
	fileStream, _, err := r.FormFile("data")
	defer fileStream.Close()

	if err != nil {
		webutils.WriteError(w, fmt.Errorf("File stream getting error: %v", err))
		return
	}

	fileSize, err := fileStream.Seek(0, os.SEEK_END)
	if err != nil {
		webutils.WriteError(w, fmt.Errorf("Cannot seek file: %v", err))
		return
	}
	fileStream.Seek(0, os.SEEK_SET)

	if f, err := vfs.DirectoryGetFile(ServerDirectory, targetFile); err != nil {
		webutils.WriteError(w, err)
	} else {
		defer f.Close()
		if err := vfs.OpenFileAndCopy(f, io.NewSectionReader(fileStream, 0, fileSize)); err != nil {
			webutils.WriteError(w, fmt.Errorf("Error when updating pack file: %v", err))
		}
	}
}
func HandlerUploadPackFileParam(w http.ResponseWriter, r *http.Request) {
	targetFile := mux.Vars(r)["file"]
	param := mux.Vars(r)["param"]
	fileStream, _, err := r.FormFile("data")
	defer fileStream.Close()
	if err != nil {
		webutils.WriteError(w, fmt.Errorf("File stream getting error: %v", err))
		return
	}

	data, err := pack.GetInstanceHandler(ServerDirectory, targetFile)
	if err != nil {
		log.Printf("Error getting instance from pack: %v", err)
		webutils.WriteError(w, err)
	} else {
		switch data.(type) {
		case *file_wad.Wad:
			wad := data.(*file_wad.Wad)
			id, err := strconv.Atoi(param)
			if err != nil {
				webutils.WriteError(w, fmt.Errorf("target wad resource name '%s' is not integer: %v", param, err))
			} else {
				if fileData, err := ioutil.ReadAll(fileStream); err == nil {
					if err := wad.UpdateTagsData(map[file_wad.TagId][]byte{file_wad.TagId(id): fileData}); err != nil {
						webutils.WriteError(w, fmt.Errorf("Error updating tags: %v", err))
					}
				} else {
					webutils.WriteError(w, fmt.Errorf("reading file error: %v", err))
				}
			}
		}
	}

}
