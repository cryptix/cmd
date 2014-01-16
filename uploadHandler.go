package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const maxSize = 1024 * 1024 * 5

func uploadHandler(resp http.ResponseWriter, req *http.Request, log *log.Logger) {
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

	io.Copy(input, file)
	log.Printf("Got File: %s\n", header.Filename)
	http.Redirect(resp, req, "/", http.StatusFound)
}
