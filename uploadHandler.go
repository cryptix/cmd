package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const maxSize = 1024 * 1024 * 5

func uploadHandler(resp http.ResponseWriter, req *http.Request) {
	file, header, err := req.FormFile("fupload")
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		fmt.Fprintf(os.Stderr, "uploadHandler - req.FormFile - Error: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "uploadHandler - Got Upload: %v\n", header.Filename)

	input, err := os.Create(filepath.Join(dumpDir, header.Filename))
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		fmt.Fprintf(os.Stderr, "uploadHandler - os.Open - Error: %v\n", err)
		return
	}
	defer input.Close()

	io.Copy(input, file)
	fmt.Fprintf(os.Stderr, "Upload Done - %s\n", header.Filename)
	http.Redirect(resp, req, "/", http.StatusFound)
}
