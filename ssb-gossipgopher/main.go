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
		//var args []map[string]interface{}
		var args []struct {
			Timeout int
		}
		err := json.Unmarshal(msg, &args)
		if err != nil {
			return errors.Wrap(err, "failed to decode ping arguments")
		}
		timeout := 1000
		if len(args) == 1 {
			/*
				if t, hasTimeout := args[0]["timeout"]; hasTimeout {
					timeout = t.(float64)
				}
			*/
			timeout = args[0].Timeout
		}
		log.Log("event", "incoming call", "cmd", "ping", "timeout", timeout)
		return struct {
			Pong string
		}{"test"}
	})

	client.HandleSource("blobs.createWants", func(msg json.RawMessage) chan interface{} {
		log.Log("event", "incoming call", "cmd", "blob wants", "msg", string(msg))
		type blobs struct {
			Ref string
		}
		blobWants := make(chan interface{})
		go func() {
			for i := 0; i <= 10; i++ {
				blobWants <- blobs{
					Ref: "123",
				}
			}
			close(blobWants)
			log.Log("event", "source done", "cmd", "blob wants")
		}()
		return blobWants
	})

	client.HandleSource("createHistoryStream", func(msg json.RawMessage) chan interface{} {
		resp := make(chan interface{})
		var args []struct {
			Id         string
			Seq        int
			Keys, Live bool
		}
		err := json.Unmarshal(msg, &args)
		if err != nil {
			go func() {
				resp <- errors.Wrap(err, "failed to decode ping arguments")
			}()
			return resp
		}

		for i, arg := range args {
			log.Log("event", "incoming call", "cmd", "histStream", "i", i, "id", arg.Id, "seq", arg.Seq)
		}
		type reply struct {
			Author string
			Msg    string
		}
		go func() {
			for i := 0; i <= 10; i++ {
				resp <- reply{
					Author: "@123",
					Msg:    fmt.Sprintf("Msg%d", i),
				}
			}
			close(resp)
			log.Log("event", "source done", "cmd", "histStream")
		}()
		return resp
	})

	//""

	return <-running
}
