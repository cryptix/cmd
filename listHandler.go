package main

import (
	"fmt"
	"net/http"
	"os"
)

func listHandler(resp http.ResponseWriter, req *http.Request) {
	dir, err := os.Open(dumpDir)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		fmt.Fprintf(os.Stderr, "listHandler - os.Open - Error: %v\n", err)
		return
	}

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		fmt.Fprintf(os.Stderr, "listHandler - dir.Readdir - Error: %v\n", err)
		return
	}

	listTemplate.Execute(resp, fileInfos)
}
