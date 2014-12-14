package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

func tmpl(a asset) *template.Template {
	return template.Must(template.New("listTemplate").Parse(a.Content))
}

func listHandler(resp http.ResponseWriter, req *http.Request) {
	dir, err := os.Open(*dumpDir)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Printf("listHandler - os.Open - Error: %v\n", err)
		return
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Printf("listHandler - dir.Readdir - Error: %v\n", err)
		return
	}

	list.Execute(resp, fileInfos)
}
