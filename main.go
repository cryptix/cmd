package main

import (
	"github.com/codegangsta/martini"
	"html/template"
)

const listTemplateText = `
<h1>Upload File</h1>
<form action="/upload" method="post" enctype="multipart/form-data">
  <input type="file" name="fupload"/>
  <input type="submit" value="Upload" />
</form>

<h1>List of files</h1>
<ul>
{{range .}}
<li><a href="/{{.Name}}">{{.Name}}</a></li>
{{end}}
</ul>

<h1>Zip of all files</h1>
<a href="/downloadAll">Download</a>
`

var listTemplate = template.Must(template.New("listTemplate").Parse(listTemplateText))

const dumpDir = "files"

func main() {
	m := martini.Classic()

	m.Use(martini.Static(dumpDir))

	m.Get("/", listHandler)
	m.Get("/downloadAll", zipDownloadHandler)
	m.Post("/upload", uploadHandler)

	m.Run()
}
