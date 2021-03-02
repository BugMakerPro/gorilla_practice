package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

func main() {
	var dir string

	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()
	r := mux.NewRouter()

	// This will serve files under http://localhost:8000/static/<filename>
	//r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))

	spa := spaHandler{staticPath: "./dist", indexPath: "index.html"}
	r.PathPrefix("/").Handler(spa)

	srv := &http.Server{
		Handler:      r,
		Addr:         "192.168.161.111:8001",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
