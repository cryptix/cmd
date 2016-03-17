package main

import (
	"html/template"
	"net/http"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/shurcooL/httpfs/html/vfstemplate"
)

var (
	tplFuncs = template.FuncMap{
		"bytes": func(s int64) string { return humanize.Bytes(uint64(s)) },
	}

	tpl = template.Must(vfstemplate.ParseGlob(assets, template.New("base").Funcs(tplFuncs), "/*.tmpl"))
)

func jsHandler(resp http.ResponseWriter, req *http.Request) error {
	dir, err := os.Open(*dumpDir)
	if err != nil {
		return err
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	return tpl.Lookup("js.tmpl").Execute(resp, fileInfos)
}

func nojsHandler(resp http.ResponseWriter, req *http.Request) error {
	return tpl.Lookup("nojs.tmpl").Execute(resp, nil)
}
