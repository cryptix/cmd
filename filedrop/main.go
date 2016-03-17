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

	"github.com/cryptix/go/http/render"
	"github.com/goji/httpauth"
	"github.com/rs/xaccess"
	"github.com/rs/xhandler"
	"github.com/rs/xlog"
	"github.com/rs/xmux"
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
)

func main() {
	flag.Parse()

	c := xhandler.Chain{}

	hostName, _ := os.Hostname()
	conf := xlog.Config{
		Fields: xlog.F{
			"role": "filedrop",
			"host": hostName,
		},
	}

	c.UseC(xlog.NewHandler(conf))

	// Plug the xlog handler's input to Go's default logger
	log.SetFlags(0)
	log.SetOutput(xlog.New(conf))

	c.UseC(xlog.RemoteAddrHandler("ip"))
	c.UseC(xlog.UserAgentHandler("user_agent"))
	c.UseC(xlog.RefererHandler("referer"))
	c.UseC(xlog.RequestIDHandler("req_id", "Request-Id"))
	c.UseC(xaccess.NewHandler())

	mux := xmux.New()
	mux.GET("/js", render.HTML(jsHandler))
	mux.GET("/nojs", render.HTML(nojsHandler))
	mux.GET("/downloadAll", render.Binary(zipDownloadHandler))
	mux.POST("/upload", xhandler.HandlerFuncC(uploadHandler))

	mux.Handle("GET", "/assets/*filepath", http.StripPrefix("/assets/", gzip_file_server.New(assets)))
	mux.Handle("GET", "/drop/*filepath", http.StripPrefix("/drop/", http.FileServer(http.Dir(*dumpDir))))

	if *user != "" {
		c.Use(httpauth.SimpleBasicAuth(*user, *pass))
	}

	if *port == "0" && os.Getenv("PORT") != "" {
		*port = os.Getenv("PORT")
	}

	l, err := net.Listen("tcp", *host+":"+*port)
	checkFatal(err)

	var server http.Server
	server.Handler = c.Handler(mux)

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
