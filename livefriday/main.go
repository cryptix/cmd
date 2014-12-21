package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/jaschaephraim/lrserver"
	"github.com/russross/blackfriday"
	"github.com/skratchdot/open-golang/open"
	"gopkg.in/fsnotify.v1"
)

var watchDir = "."

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
	<!-- <link href="http://kevinburke.bitbucket.org/markdowncss/markdown.css" rel="stylesheet"></link>-->
	<script src="http://localhost:35729/livereload.js"></script>
<body>
{{.}}
</body>
</html>`))

func main() {
	app := cli.NewApp()
	app.Name = "livefriday"
	app.Usage = "see your markdown grow as you save it"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "dir,d", Value: ".", Usage: "The directory to watch and compile"},
		cli.StringFlag{Name: "host", Value: "localhost", Usage: "The http host to listen on"},
		cli.IntFlag{Name: "port,p", Value: 3000, Usage: "The http port to listen on"},
	}
	app.Action = run

	app.Run(os.Args)

}

func run(c *cli.Context) {
	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	check(err)
	defer watcher.Close()

	watchDir = c.String("dir")

	// Add dir to watcher
	err = watcher.Add(watchDir)
	check(err)

	// Start LiveReload server
	go lrserver.ListenAndServe()

	// Start goroutine that requests reload upon watcher event
	go func() {
		for {
			event := <-watcher.Events
			if strings.HasSuffix(event.Name, ".md") || strings.HasSuffix(event.Name, ".mdown") {
				lrserver.Reload(event.Name)
			}
		}
	}()

	// Start serving html
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/md", mdHandler)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir(watchDir))))

	listenAddr := fmt.Sprintf("%s:%d", c.String("host"), c.Int("port"))

	done := make(chan struct{})
	go func() {
		err = http.ListenAndServe(listenAddr, nil)
		check(err)
		close(done)
	}()

	err = open.Run("http://" + listenAddr)
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
		if strings.HasSuffix(n, ".md") || strings.HasSuffix(n, ".mdown") {
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

	input, err := ioutil.ReadFile(filepath.Join(watchDir, filepath.Base(fname)))
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
