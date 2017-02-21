package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/errgo.v1"
)

func uploadHandler(resp http.ResponseWriter, req *http.Request) {
	file, header, err := req.FormFile("fupload")
	if err != nil {
		err = errgo.Notef(err, "uploadHandler: req.FormFile failed")
		log.Log("func", "uploadHandler", "error", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	input, err := os.Create(filepath.Join(*dumpDir, header.Filename))
	if err != nil {
		err = errgo.Notef(err, "uploadHandler: os.Open failed")
		log.Log("func", "uploadHandler", "error", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	defer input.Close()

	n, err := io.Copy(input, file)
	if err != nil {
		err = errgo.Notef(err, "uploadHandler: io.Copy failed")
		log.Log("func", "uploadHandler", "error", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Log("event", "upload complete", "name", header.Filename, "size", n)
	fmt.Fprintf(resp, `{"data":"%s", "status":%d}`, "upload complete", http.StatusCreated)
}
