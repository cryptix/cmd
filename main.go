package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jaschaephraim/lrserver"
	"github.com/russross/blackfriday"
	"github.com/skratchdot/open-golang/open"
	"gopkg.in/fsnotify.v0"
)

const watchDir = "."

// template with list of files found in watchDir (with livereload)
var mdList = template.Must(template.New("mdList").Parse(`<!doctype html>
<html>
<head>
	<title>List of Markdown files</title>
	<script src="http://localhost:35729/livereload.js"></script>
<body>
<ul>
{{range .}}
	<li><a href="/md?file={{.}}">{{.}}</a></li>
{{end}}
</ul>
</body>
</html>`))

// template for rendering markdown content (with livereload)
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
	check(err)
	defer watcher.Close()

	// Add dir to watcher
	err = watcher.Add(watchDir)
	check(err)

	// Start LiveReload server
	go lrserver.ListenAndServe()

	// Start goroutine that requests reload upon watcher event
	go func() {
		for {
			event := <-watcher.Events
			if strings.HasSuffix(event.Name, ".md") {
				lrserver.Reload(event.Name)
			}
		}
	}()

	// Start serving html
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/md", mdHandler)

	done := make(chan struct{})
	go func() {
		err = http.ListenAndServe(":3000", nil)
		check(err)
		close(done)
	}()

	err = open.Run("http://localhost:3000")
	check(err)

	<-done
}

// indexHandler builds a list with links to all .md files in the watchDir
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

// mdHandler ReadFile's the passed file argument and puts it through blackfriday
func mdHandler(rw http.ResponseWriter, req *http.Request) {
	fname := req.URL.Query().Get("file")
	if fname == "" {
		http.Error(rw, "no fname", http.StatusBadRequest)
		return
	}

	input, err := ioutil.ReadFile(filepath.Base(fname))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	md.Execute(rw, template.HTML(blackfriday.MarkdownCommon(input)))
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
