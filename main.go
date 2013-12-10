package main

import (
	"github.com/codegangsta/martini"
)

const dumpDir = "files"

func main() {
	m := martini.Classic()

	m.Use(martini.Static(dumpDir))

	m.Get("/", listHandler)
	m.Get("/downloadAll", zipDownloadHandler)
	m.Post("/upload", uploadHandler)

	m.Run()
}
