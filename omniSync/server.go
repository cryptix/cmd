package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/cloudflare/cfssl/transport"
	"github.com/cloudflare/cfssl/transport/core"
	"github.com/cloudflare/cfssl/transport/example/exlib"
	"gopkg.in/errgo.v1"

	"golang.org/x/net/webdav"
)

var (
	addr  = flag.String("addr", "acab.mobi:4444", "address of the server")
	dir   = flag.String("dir", "./davOne", "where to save the files")
	cfCfg = flag.String("cfgFile", "cfssl_config.json", "cloudflare ssl config")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	lis, err := parseAndSetupListener(*addr, *cfCfg)
	if err != nil {
		log.Fatal("parseAndSetupListen:", err)
	}

	h := &webdav.Handler{
		FileSystem: webdav.Dir(*dir),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {

			switch r.Method {
			case "COPY", "MOVE":
				dst := ""
				if u, err := url.Parse(r.Header.Get("Destination")); err == nil {
					dst = u.Path
				}
				o := r.Header.Get("Overwrite")
				log.Printf("%-10s%-30s%-30so=%-2s%v", r.Method, r.URL.Path, dst, o, err)
			default:
				log.Printf("%-10s%-30s%v", r.Method, r.URL.Path, err)
			}
		},
	}

	http.Handle("/", h)

	cfg, err := lis.TLSClientAuthServerConfig()
	if err != nil {
		exlib.Err(1, err, "tlsconfig")
	}

	server := &http.Server{
		Addr:      ":https",
		TLSConfig: cfg,
	}

	log.Fatal(server.Serve(lis))
}

func parseAndSetupListener(addr, cfg string) (*transport.Listener, error) {
	f, err := os.Open(cfg)
	if err != nil {
		return nil, errgo.Notef(err, "opening config file failed")
	}

	var id = new(core.Identity)
	err = json.NewDecoder(f).Decode(id)
	if err != nil {
		return nil, errgo.Notef(err, "parsing config json failed")
	}

	tr, err := transport.New(exlib.Before, id)
	if err != nil {
		return nil, errgo.Notef(err, "parsing config json failed")
	}

	l, err := transport.Listen(addr, tr)
	if err != nil {
		return nil, errgo.Notef(err, "failed to create listener")
	}

	var errChan = make(chan error, 0)
	go func(ec <-chan error) {
		for {
			err, ok := <-ec
			if !ok {
				log.Println("error channel closed, future errors will not be reported")
				break
			}
			log.Fatal("auto update error: %v", err)
		}
	}(errChan)

	log.Print("setting up auto-update")
	go l.AutoUpdate(nil, errChan)

	return l, nil
}
