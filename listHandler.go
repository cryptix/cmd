package main

import (
	"html/template"
	"net/http"
	"os"
)

var listTmpl = template.Must(template.New("listTemplate").Parse(assetMustString("list.tmpl")))

func listHandler(resp http.ResponseWriter, req *http.Request) error {
	dir, err := os.Open(*dumpDir)
	if err != nil {
		return err
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	return listTmpl.Execute(resp, fileInfos)
}
