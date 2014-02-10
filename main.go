package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/codegangsta/martini"
)

var (
	host    = flag.String("host", "localhost", "The hostname/ip to listen on.")
	port    = flag.Int("port", 3000, "The port number to listen on.")
	dumpDir = flag.String("dir", "files", "The directory used to store and serve files")
)

func main() {
	flag.Parse()

	m := martini.Classic()

	m.Use(martini.Static(*dumpDir))

	m.Get("/", listHandler)
	m.Get("/downloadAll", zipDownloadHandler)
	m.Post("/upload", uploadHandler)

	http.ListenAndServe(fmt.Sprintf("%s:%d", *host, +*port), m)
}
