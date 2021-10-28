package webutils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
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

func WriteJsonFile(w http.ResponseWriter, v interface{}, fileName string) {
	if data, err := json.MarshalIndent(v, "", "  "); err != nil {
		WriteError(w, errors.Wrapf(err, "Failed to marshal"))
	} else {
		WriteFile(w, bytes.NewReader(data), fileName+".json")
	}
}

func ReadJsonFile(r *http.Request, formFileKey string, v interface{}) error {
	if strings.ToUpper(r.Method) != "POST" {
		return errors.Errorf("Invalid http method %q", r.Method)
	}

	f, _, err := r.FormFile(formFileKey)
	if err != nil {
		return errors.Wrapf(err, "Failed to get file")
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.Wrapf(err, "Failed to read")
	}

	if err := json.Unmarshal(data, v); err != nil {
		return errors.Wrapf(err, "Failed to unmarshal")
	}

	return nil
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
