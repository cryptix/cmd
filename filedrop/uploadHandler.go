package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func uploadHandler(resp http.ResponseWriter, req *http.Request) {
	file, header, err := req.FormFile("fupload")
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Printf("uploadHandler - req.FormFile - Error: %v\n", err)
		return
	}

	input, err := os.Create(filepath.Join(*dumpDir, header.Filename))
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Printf("uploadHandler - os.Open - Error: %v\n", err)
		return
	}
	defer input.Close()

	n, err := io.Copy(input, file)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Printf("uploadHandler - io.Copy - Error: %v\n", err)
		return
	}

	log.Printf("Got File: %s Len: %d\n", header.Filename, n)
	fmt.Fprintf(resp, `{"data":"%s", "status":%d}`, "upload complete", http.StatusCreated)
}
