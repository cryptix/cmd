package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/xlog"
)

func uploadHandler(resp http.ResponseWriter, req *http.Request) {
	l := xlog.FromContext(req.Context())
	l.SetField("handler", "upload")

	file, header, err := req.FormFile("fupload")
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		l.Error("req.FormFile failed.", xlog.F{"err": err})
		return
	}

	input, err := os.Create(filepath.Join(*dumpDir, header.Filename))
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		l.Error("os.Open failed.", xlog.F{"err": err})
		return
	}
	defer input.Close()

	n, err := io.Copy(input, file)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Printf("uploadHandler - io.Copy - Error: %v\n", err)
		return
	}

	l.Info("Upload complete", xlog.F{"name": header.Filename, "size": n})
	fmt.Fprintf(resp, `{"data":"%s", "status":%d}`, "upload complete", http.StatusCreated)
}
