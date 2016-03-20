/*
filedrop is a simple http server with an upload html page

- supports HTTP Basic out (-user / -pass flags)
- supports HTTPS listening (-key / -crt flags)

*/
package main

import (
	"crypto/tls"
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/cryptix/go/http/render"
	"github.com/cryptix/go/logging"
	"github.com/dkumor/acmewrapper"
	"github.com/dustin/go-humanize"
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

	ssl    = flag.Bool("ssl", false, "enable SSL (lets encrypt)")
	sslKey = flag.String("key", "key.pem", "Key-file for SSL connections")
	sslCrt = flag.String("crt", "cert.pem", "Certificate for SSL connections")
)

func main() {
	flag.Parse()

	c := xhandler.Chain{}

	hostName, _ := os.Hostname()
	conf := xlog.Config{
		Output: xlog.NewConsoleOutput(),
		Fields: xlog.F{
			"role": "filedrop",
			"host": hostName,
		},
	}
	logging.SetupLogging(&conf)

	c.UseC(xlog.NewHandler(conf))

	c.UseC(xlog.RemoteAddrHandler("ip"))
	c.UseC(xlog.UserAgentHandler("user_agent"))
	c.UseC(xlog.RefererHandler("referer"))
	c.UseC(xlog.RequestIDHandler("req_id", "Request-Id"))
	c.UseC(xaccess.NewHandler())
	if *user != "" {
		c.Use(httpauth.SimpleBasicAuth(*user, *pass))
	}

	ren, err := render.New(assets, "base.tmpl",
		render.AddTemplates("js.tmpl", "nojs.tmpl", "base.tmpl"),
		render.FuncMap(template.FuncMap{
			"bytes": func(s int64) string { return humanize.Bytes(uint64(s)) },
		}),
	)
	checkFatal(err)

	if !production {
		c.UseC(ren.GetReloader())
	}

	mux := xmux.New()
	mux.GET("/", ren.StaticHTML("base.tmpl"))
	mux.GET("/js", ren.HTML("js.tmpl", jsHandler))
	mux.GET("/nojs", ren.StaticHTML("nojs.tmpl"))
	mux.GET("/downloadAll", render.Binary(zipDownloadHandler))
	mux.POST("/upload", xhandler.HandlerFuncC(uploadHandler))

	mux.Handle("GET", "/assets/*filepath", http.StripPrefix("/assets/", gzip_file_server.New(assets)))
	mux.Handle("GET", "/drop/*filepath", http.StripPrefix("/drop/", http.FileServer(http.Dir(*dumpDir))))

	var server http.Server
	server.Handler = c.Handler(mux)

	var l net.Listener
	if *ssl {
		w, err := acmewrapper.New(acmewrapper.Config{
			Domains: []string{*host},
			Address: ":443",

			TLSCertFile: *sslCrt,
			TLSKeyFile:  *sslKey,

			// Let's Encrypt stuff
			RegistrationFile: "user.reg",
			PrivateKeyFile:   "user.pem",

			TOSCallback: acmewrapper.TOSAgree,
		})
		checkFatal(err)
		l, err = tls.Listen("tcp", ":443", w.TLSConfig())
		checkFatal(err)
		server.Addr = ":443"
		server.TLSConfig = w.TLSConfig()
		log.Printf("Serving at https://%s/", l.Addr())
	} else {
		if *port == "0" && os.Getenv("PORT") != "" {
			*port = os.Getenv("PORT")
		}

		l, err = net.Listen("tcp", *host+":"+*port)
		checkFatal(err)
		log.Printf("Serving at http://%s/", l.Addr())
	}

	checkFatal(server.Serve(l))
}

func checkFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
