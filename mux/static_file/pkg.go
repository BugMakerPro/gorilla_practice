package main

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"google.golang.org/grpc/status"

	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func muxVars(r *http.Request, key string) string {
	if v, ok := mux.Vars(r)[key]; ok {
		return v
	}
	return ""
}

func muxVarsInt64(r *http.Request, key string) int64 {
	if v, ok := mux.Vars(r)[key]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func muxVarsInt32(r *http.Request, key string) int32 {
	return int32(muxVarsInt64(r, key))
}
func muxVarsBool(r *http.Request, key string) bool {
	if int32(muxVarsInt64(r, key)) > 0 {
		return true
	} else if strings.ToLower(muxVars(r, key)) == "true" {
		return true
	}
	return false
}

// json保留空值 https://github.com/golang/protobuf/issues/799
// 然而，会把timestamp转换成字符串
func writeJson(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		writeServerError(w, "write json err=%v, data=%v", err, body)
	}
}

func writeServerError(w http.ResponseWriter, format string, err error, a ...interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	if st, ok := status.FromError(err); ok {
		err = errors.New(st.Message())
	}
	w.Write([]byte(err.Error()))
	glog.Error(fmt.Sprintf(format, append([]interface{}{err}, a...)...))
}

func writeServerRawError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	if st, ok := status.FromError(err); ok {
		err = errors.New(st.Message())
	}
	w.Write([]byte(err.Error()))
	glog.Error(err.Error())
}

func writeClientError(w http.ResponseWriter, format string, err error, a ...interface{}) {
	w.WriteHeader(http.StatusBadRequest)
	if st, ok := status.FromError(err); ok {
		err = errors.New(st.Message())
	}
	w.Write([]byte(err.Error()))
	glog.Error(fmt.Sprintf(format, append([]interface{}{err}, a...)...))
}

func redirect(w http.ResponseWriter, url string) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`window.location.href = '%v'`, url)))
}

func readJson(r *http.Request, body interface{}) error {
	return json.NewDecoder(r.Body).Decode(body)
}

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("path: ",path)
	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)
	fmt.Println("path: ",path)
	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}
