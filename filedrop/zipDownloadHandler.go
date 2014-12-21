package main

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func zipDownloadHandler(resp http.ResponseWriter, req *http.Request) error {
	dir, err := os.Open(*dumpDir)
	if err != nil {
		return err
	}

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	resp.Header().Set("Content-Disposition", "attachment; filename=files.zip")
	resp.WriteHeader(http.StatusOK)

	zipWriter := zip.NewWriter(resp)

	for _, fInfo := range fileInfos {
		if fInfo.IsDir() {
			continue
		}

		rawFile, err := os.Open(filepath.Join(*dumpDir, fInfo.Name()))
		if err != nil {
			return err
		}
		defer rawFile.Close()

		zipFile, err := zipWriter.Create(fInfo.Name())
		if err != nil {
			return err
		}

		io.Copy(zipFile, rawFile)
	}

	return zipWriter.Close()
}
