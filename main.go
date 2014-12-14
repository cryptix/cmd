package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/goji/httpauth"
)

var (
	host = flag.String("host", "localhost", "The hostname/ip to listen on.")
	port = flag.Int("port", 0, "The port number to listen on.")

	dumpDir = flag.String("dir", ".", "The directory used to store and serve files")

	user = flag.String("user", "", "HTTP BasicAuth User")
	pass = flag.String("pass", "ChangeMe", "HTTP BasicAuth User")
)

//go:generate -command asset go run asset.go
//go:generate asset list.tmpl
//go:generate asset uploadui.js
//go:generate asset bootstrapProgressbar.min.js

type JS struct {
	asset
}

func js(a asset) JS {
	return JS{a}
}

func main() {
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/list", listHandler)
	mux.HandleFunc("/downloadAll", zipDownloadHandler)
	mux.HandleFunc("/upload", uploadHandler)
	mux.Handle("/uploadui.js", uploadui)
	mux.Handle("/bootstrapProgressbar.js", bootstrapProgressbar)
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(*dumpDir))))

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	if *user != "" {
		n.UseHandler(httpauth.SimpleBasicAuth(*user, *pass)(mux))
	} else {
		n.UseHandler(mux)
	}

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Serving at http://%s/", l.Addr())

	if err := http.Serve(l, n); err != nil {
		log.Fatal(err)
	}
}
