package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
)

var (
	host    = flag.String("host", "", "The hostname/ip to listen on.")
	port    = flag.Int("port", 3000, "The port number to listen on.")
	dumpDir = flag.String("dir", ".", "The directory used to store and serve files")
)

func main() {
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/", listHandler)
	mux.HandleFunc("/downloadAll", zipDownloadHandler)
	mux.HandleFunc("/upload", uploadHandler)

	n := negroni.Classic()
	n.Use(negroni.NewStatic(http.Dir(*dumpDir)))
	n.UseHandler(mux)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	n.Run(addr)

}
