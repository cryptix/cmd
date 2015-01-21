//go:generate go-bindata -pkg=$GOPACKAGE -prefix=assets assets/...

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/cryptix/go/http/render"
	"github.com/goji/httpauth"
)

var (
	host = flag.String("host", "localhost", "The hostname/ip to listen on.")
	port = flag.String("port", "0", "The port number to listen on.")

	dumpDir = flag.String("dir", ".", "The directory used to store and serve files")

	user = flag.String("user", "", "HTTP BasicAuth User")
	pass = flag.String("pass", "ChangeMe", "HTTP BasicAuth User")

	progStart = time.Now()
)

func main() {
	flag.Parse()

	mux := http.NewServeMux()

	mux.Handle("/list", render.HTML(listHandler))
	mux.Handle("/downloadAll", render.Binary(zipDownloadHandler))
	mux.HandleFunc("/upload", uploadHandler)

	mux.Handle("/uploadui.js", render.Binary(serveAsset))
	mux.Handle("/bootstrapProgressbar.min.js", render.Binary(serveAsset))

	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(*dumpDir))))

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	if *user != "" {
		n.UseHandler(httpauth.SimpleBasicAuth(*user, *pass)(mux))
	} else {
		n.UseHandler(mux)
	}

	if *port == "0" {
		*port = os.Getenv("PORT")
	}

	l, err := net.Listen("tcp", *host+":"+*port)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Serving at http://%s/", l.Addr())

	if err := http.Serve(l, n); err != nil {
		log.Fatal(err)
	}
}

func serveAsset(w http.ResponseWriter, req *http.Request) error {
	b, err := Asset(req.URL.Path[1:])
	if err != nil {
		return err
	}

	http.ServeContent(w, req, req.URL.Path[1:], progStart, bytes.NewReader(b))
	return nil
}

func assetMustString(name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	f, ok := _bindata[cannonicalName]
	if !ok {
		log.Fatal(fmt.Errorf("Asset %s not found", name))
		return ""
	}

	b, err := f()
	if err != nil {
		log.Fatal(err)
	}
	return string(b)

}
