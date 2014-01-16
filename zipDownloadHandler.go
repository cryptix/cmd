package main

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func zipDownloadHandler(resp http.ResponseWriter, req *http.Request, log *log.Logger) {
	dir, err := os.Open(*dumpDir)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Printf("listHandler - os.Open(dumpDir) - Error: %v\n", err)
		return
	}

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Printf("listHandler - dir.Readdir - Error: %v\n", err)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Header().Set("Content-Description", "File Transfer")
	resp.Header().Set("Content-type", "application/octet-stream")
	resp.Header().Set("Content-Disposition", "attachment; filename=files.zip")
	resp.Header().Set("Content-Transfer-Encoding", "binary")

	zipWriter := zip.NewWriter(resp)

	for _, fInfo := range fileInfos {
		if fInfo.IsDir() {
			continue
		}

		rawFile, err := os.Open(filepath.Join(*dumpDir, fInfo.Name()))
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			log.Printf("listHandler - os.Open(zipFile) - Error: %v\n", err)
			return
		}
		defer rawFile.Close()

		zipFile, err := zipWriter.Create(fInfo.Name())
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			log.Printf("listHandler - zipWriter.Create - Error: %v\n", err)
			return
		}

		io.Copy(zipFile, rawFile)
	}

	err = zipWriter.Close()
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Printf("listHandler - zipWriter.Close - Error: %v\n", err)
		return
	}

	log.Println("Served .zip File")
}
