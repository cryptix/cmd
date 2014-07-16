package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/russross/blackfriday"

	"github.com/jaschaephraim/lrserver"
	"gopkg.in/fsnotify.v0"
)

const watchDir = "."

// html includes the client JavaScript
var mdList = template.Must(template.New("mdList").Parse(`<!doctype html>
<html>
<head>
	<title>Example</title>
	<script src="http://localhost:35729/livereload.js"></script>
<body>
<ul>
{{range .}}
	<li><a href="/md?file={{.}}">{{.}}</a></li>
{{end}}
</ul>
</body>
</html>`))

var md = template.Must(template.New("md").Parse(`<!doctype html>
<html>
<head>
	<script src="http://localhost:35729/livereload.js"></script>
<body>
{{.}}
</body>
</html>`))

func main() {
	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer watcher.Close()

	// Add dir to watcher
	err = watcher.Add(watchDir)
	if err != nil {
		log.Fatalln(err)
	}

	// Start LiveReload server
	go lrserver.ListenAndServe()

	// Start goroutine that requests reload upon watcher event
	go func() {
		for {
			event := <-watcher.Events
			if strings.HasSuffix(event.Name, ".md") {
				fmt.Println("Realoading:", event.Name)
				lrserver.Reload(event.Name)
			}
		}
	}()

	// Start serving html
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/md", mdHandler)
	http.ListenAndServe(":3000", nil)
}

func indexHandler(rw http.ResponseWriter, req *http.Request) {
	dir, err := os.Open(watchDir)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	dirNames, err := dir.Readdirnames(-1)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	mdFiles := make([]string, len(dirNames))
	i := 0
	for _, n := range dirNames {
		if strings.HasSuffix(n, ".md") {
			mdFiles[i] = n
			i++
		}
	}

	rw.WriteHeader(http.StatusOK)
	mdList.Execute(rw, mdFiles[:i])
}

func mdHandler(rw http.ResponseWriter, req *http.Request) {
	fname := req.URL.Query().Get("file")
	if fname == "" {
		http.Error(rw, "no fname", http.StatusBadRequest)
		return
	}

	input, err := ioutil.ReadFile(fname)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	md.Execute(rw, template.HTML(blackfriday.MarkdownCommon(input)))
}
