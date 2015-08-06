/*
filedrop is a simple http server with an upload html page

- supports HTTP Basic out (-user / -pass flags)
- supports HTTPS listening (-key / -crt flags)

*/
package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/cryptix/go/http/render"
	"github.com/goji/httpauth"
	"github.com/shurcooL/go/gzip_file_server"
)

var (
	host = flag.String("host", "localhost", "The hostname/ip to listen on.")
	port = flag.String("port", "0", "The port number to listen on.")

	dumpDir = flag.String("dir", ".", "The directory used to store and serve files")

	user = flag.String("user", "", "HTTP BasicAuth User")
	pass = flag.String("pass", "ChangeMe", "HTTP BasicAuth User")

	sslKey = flag.String("key", "", "Key-file for SSL connections")
	sslCrt = flag.String("crt", "", "Certificate for SSL connections")

	progStart = time.Now()
)

// Redirect to public ipfs gateways - get your browser plugin..!
type ipfsRedirectHandler struct {
	Gateway string
}

func (rh *ipfsRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !(strings.HasPrefix(r.URL.Path, "/ipfs/") || strings.HasPrefix(r.URL.Path, "/ipns/")) {
		log.Println("Path:", r.URL.Path)
		http.Error(w, "ipfsHandler: unknown path prefix", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, rh.Gateway+r.URL.Path, http.StatusFound)
}

func main() {
	flag.Parse()

	mux := http.NewServeMux()

	mux.Handle("/list", render.HTML(listHandler))
	mux.Handle("/downloadAll", render.Binary(zipDownloadHandler))
	mux.HandleFunc("/upload", uploadHandler)

	irh := &ipfsRedirectHandler{Gateway: "http://gateway.ipfs.io"}
	mux.Handle("/ipfs/", irh)
	mux.Handle("/ipns/", irh)

	mux.Handle("/assets/", http.StripPrefix("/assets/", gzip_file_server.New(assets)))
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
	checkFatal(err)

	var server http.Server
	server.Handler = n

	if *sslKey != "" {
		tlsCfg := &tls.Config{}
		if tlsCfg.NextProtos == nil {
			tlsCfg.NextProtos = []string{"http/1.1"}
		}
		var err error
		tlsCfg.Certificates = make([]tls.Certificate, 1)
		tlsCfg.Certificates[0], err = tls.LoadX509KeyPair(*sslCrt, *sslKey)
		checkFatal(err)

		l = tls.NewListener(l, tlsCfg)
		log.Printf("Serving at https://%s/", l.Addr())
	} else {
		log.Printf("Serving at http://%s/", l.Addr())
	}

	checkFatal(server.Serve(l))
}

func checkFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
