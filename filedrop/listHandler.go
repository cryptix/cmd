package main

import (
	"net/http"
	"os"

	"golang.org/x/net/context"
)

func jsHandler(ctx context.Context, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	dir, err := os.Open(*dumpDir)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	return fileInfos, nil
}
