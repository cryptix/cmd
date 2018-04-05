package main

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/cryptix/go/logging"
)

var check = logging.CheckFatal

func main() {
	f, err := os.Create(os.Args[1])
	check(err)
	defer f.Close()

	zw := zip.NewWriter(f)

	err = filepath.Walk(os.Args[2], func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			w, err := zw.Create(p)
			check(err)

			zf, err := os.Open(p)
			check(err)

			_, err = io.Copy(w, zf)
			check(err)

			log.Println("copied ", p)
		}
		return nil
	})
	check(err)

	check(zw.Close())
	log.Println("zip written..!")
}
