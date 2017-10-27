package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"

	"github.com/cryptix/go/logging"
	"github.com/cryptix/secretstream"
	"github.com/cryptix/secretstream/secrethandshake"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v2"
	"scuttlebot.io/go/muxrpc"
	"scuttlebot.io/go/muxrpc/codec"
)

var (
	sbotAppKey     []byte
	defaultKeyFile string
	client         *muxrpc.Client

	running chan error

	log   logging.Interface
	check = logging.CheckFatal

	Revision = "unset"
)

func init() {
	var err error
	sbotAppKey, err = base64.StdEncoding.DecodeString("1KHLiKZvAvjbY1ziZEHMXawbCEIM6qwjCDm3VYRan/s=")
	check(err)

	u, err := user.Current()
	check(err)

	defaultKeyFile = filepath.Join(u.HomeDir, ".ssb", "secret")
}

func main() {
	logging.SetupLogging(nil)
	log = logging.Logger("gossipgopher")

	app := cli.App{
		Name:    "ssb-gossipgopher",
		Usage:   "very chatty hermit gopher",
		Version: "alpha1",
	}
	cli.VersionPrinter = func(c *cli.Context) {
		// go install -ldflags="-X main.Revision=$(git rev-parse HEAD)"
		fmt.Printf("%s ( rev: %s )\n", c.App.Version, Revision)
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "addr", Value: ":8008", Usage: "tcp address to listen on"},
		&cli.StringFlag{Name: "key,k", Value: defaultKeyFile},
		&cli.BoolFlag{Name: "verbose,vv", Usage: "print muxrpc packets"},
	}
	app.Before = initClient
	app.Commands = []*cli.Command{
		{
			Name:   "serve",
			Action: serveCmd,
		},
	}

	check(app.Run(os.Args))
}

func initClient(ctx *cli.Context) error {
	localKey, err := secrethandshake.LoadSSBKeyPair(ctx.String("key"))
	if err != nil {
		return errors.Wrap(err, "failed to load keypair")
	}
	var conn net.Conn
	srv, err := secretstream.NewServer(*localKey, sbotAppKey)
	if err != nil {
		return errors.Wrap(err, "failed to create server")
	}
	l, err := srv.Listen("tcp", ctx.String("addr"))
	if err != nil {
		return err
	}
	log.Log("msg", "listener created", "addr", l.Addr().String())

	conn, err = l.Accept()
	if err != nil {
		return err
	}
	log.Log("msg", "remote connection accepted", "addr", conn.RemoteAddr().String())

	if ctx.Bool("verbose") {
		client = muxrpc.NewClient(log, codec.Wrap(log, conn))
	} else {
		client = muxrpc.NewClient(log, conn)
	}

	running = make(chan error)
	go func() {
		client.Handle()
		log.Log("warning", "muxrpc disconnected")
		running <- errors.New("muxrpc closed")
	}()
	return nil
}

func serveCmd(ctx *cli.Context) error {
	client.HandleCall("gossip.ping", func(msg json.RawMessage) interface{} {
		return nil
	})

	client.HandleSource("blobs.createWants", func(msg json.RawMessage) chan interface{} {
		return nil
	})

	client.HandleSource("createHistoryStream", func(msg json.RawMessage) chan interface{} {
		return nil
	})

	//""

	return <-running
}
