package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

const listTemplateText = `
<h1>Upload File</h1>
<form action="/upload" method="post" enctype="multipart/form-data">
  <input type="file" name="fupload"/>
  <input type="submit" value="Upload" />
</form>

<h1>List of files</h1>
<table>
<thead>
	<tr>
		<td>Name</td>
		<td>Size</td>
	</tr>
</thead>
<tbody>
{{range .}}
<tr>
	<td><a href="/{{.Name}}">{{.Name}}</a></td>
	<td>{{.Size}} Bytes</td>
</tr>
{{end}}
</tbody>
</table>

<h1>Zip of all files</h1>
<a href="/downloadAll">Download</a>
`

var listTemplate = template.Must(template.New("listTemplate").Parse(listTemplateText))

func listHandler(resp http.ResponseWriter, req *http.Request, log *log.Logger) {
	dir, err := os.Open(dumpDir)
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

	listTemplate.Execute(resp, fileInfos)
}
