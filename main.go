package main

import (
	"flag"
	"github.com/codegangsta/martini"
)

var dumpDir = flag.String("dir", "files", "The directory used to store and serve files")

func main() {
	flag.Parse()

	m := martini.Classic()

	m.Use(martini.Static(*dumpDir))

	m.Get("/", listHandler)
	m.Get("/downloadAll", zipDownloadHandler)
	m.Post("/upload", uploadHandler)

	m.Run()
}
