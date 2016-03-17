package main

import (
	"html/template"
	"net/http"
	"os"

	"github.com/shurcooL/httpfs/html/vfstemplate"
)

var (
	jsTmpl   = template.Must(vfstemplate.ParseFiles(assets, nil, "/js.tmpl"))
	nojsTmpl = template.Must(vfstemplate.ParseFiles(assets, nil, "/nojs.tmpl"))
)

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

	return jsTmpl.Execute(resp, fileInfos)
}

func nojsHandler(resp http.ResponseWriter, req *http.Request) error {
	return nojsTmpl.Execute(resp, nil)
}
