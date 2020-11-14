package webutils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func WriteFileHeaders(w http.ResponseWriter, name string) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
}

func WriteFile(w http.ResponseWriter, in io.Reader, name string) {
	WriteFileHeaders(w, name)
	io.Copy(w, in)
}

func WriteJson(w http.ResponseWriter, data interface{}) {
	res, err := json.Marshal(data)
	if err != nil {
		WriteError(w, err)
	} else {
		WriteResult(w, res)
	}
}

func WriteResult(w http.ResponseWriter, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		log.Printf("Error when writing response: %v", err)
	}
}

func WriteError(w http.ResponseWriter, err error) {
	type jError struct {
		Error string `json:"error"`
	}
	data, err := json.Marshal(&jError{Error: err.Error()})
	if err == nil {
		log.Printf("HERR: %v", string(data))
		WriteResult(w, data)
	} else {
		log.Printf("Error marshaling error '%v': %v", err, data)
	}
}
