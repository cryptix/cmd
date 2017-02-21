/*
filedrop is a simple http server with an upload html page

- supports HTTP Basic out (-user / -pass flags)
- supports HTTPS listening (-key / -crt flags)

*/
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"

	"github.com/cryptix/go/http/render"
	"github.com/cryptix/go/logging"
	"github.com/dkumor/acmewrapper"
	"github.com/dustin/go-humanize"
	kitlog "github.com/go-kit/kit/log"
	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/xaccess"
	"github.com/rs/xlog"
	"github.com/shurcooL/httpgzip"
)

var log *kitlog.Context

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

	logging.SetupLogging(nil)
	log = logging.Logger("filedrop")

	chain := alice.New(
		// TODO: move to new logging
		xlog.RemoteAddrHandler("ip"),
		xlog.UserAgentHandler("user_agent"),
		xlog.RefererHandler("referer"),
		xlog.RequestIDHandler("req_id", "Request-Id"),
		xaccess.NewHandler())
	/*
		c.Use(xlog.NewHandler(conf))
	*/
	if *user != "" {
		chain = chain.Append(httpauth.SimpleBasicAuth(*user, *pass))
	}

	ren, err := render.New(assets,
		render.BaseTemplate("base.tmpl"),
		render.AddTemplates("js.tmpl", "nojs.tmpl", "base.tmpl"),
		render.FuncMap(template.FuncMap{
			"bytes": func(s int64) string { return humanize.Bytes(uint64(s)) },
		}),
	)
	logging.CheckFatal(err)

	if !production {
		chain = chain.Append(ren.GetReloader())
	}

	mux := mux.NewRouter()
	mux.Handle("/", ren.StaticHTML("base.tmpl"))
	mux.HandleFunc("/js", ren.HTML("js.tmpl", jsHandler))
	mux.Handle("/nojs", ren.StaticHTML("nojs.tmpl"))
	mux.Handle("/downloadAll", render.Binary(zipDownloadHandler))
	mux.HandleFunc("/upload", uploadHandler).Methods("POST")

	mux.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", httpgzip.FileServer(assets, httpgzip.FileServerOptions{})))
	mux.PathPrefix("/drop/").Handler(http.StripPrefix("/drop/", http.FileServer(http.Dir(*dumpDir))))

	var server http.Server
	server.Handler = chain.Then(mux)

	var l net.Listener
	var lisAddr string
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
		logging.CheckFatal(err)
		l, err = tls.Listen("tcp", ":443", w.TLSConfig())
		logging.CheckFatal(err)
		server.Addr = ":443"
		server.TLSConfig = w.TLSConfig()
		lisAddr = fmt.Sprintf("https://%s/", l.Addr())

	} else {
		if *port == "0" && os.Getenv("PORT") != "" {
			*port = os.Getenv("PORT")
		}

		l, err = net.Listen("tcp", *host+":"+*port)
		logging.CheckFatal(err)
		lisAddr = fmt.Sprintf("http://%s/", l.Addr())
	}
	log.Log("event", "serving", "addr", lisAddr)
	logging.CheckFatal(server.Serve(l))
}
