package main

import (
	"log"
	"net/http"
	"os"
)

func listHandler(resp http.ResponseWriter, req *http.Request, log *log.Logger) {
	dir, err := os.Open(dumpDir)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Printf("listHandler - os.Open - Error: %v\n", err)
		return
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Printf("listHandler - dir.Readdir - Error: %v\n", err)
		return
	}

	listTemplate.Execute(resp, fileInfos)
}
