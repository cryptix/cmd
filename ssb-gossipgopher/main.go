package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"cryptoscope.co/go/muxrpc"
	"cryptoscope.co/go/muxrpc/codec"
	"cryptoscope.co/go/secretstream"
	"cryptoscope.co/go/secretstream/secrethandshake"
	"github.com/cryptix/go/debug"
	"github.com/cryptix/go/logging"
	humanize "github.com/dustin/go-humanize"
	kitlog "github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	cli "gopkg.in/urfave/cli.v2"
)

var (
	sbotAppKey     []byte
	defaultKeyFile string
	packer         muxrpc.Packer
	rpc            muxrpc.Endpoint
	localKey       *secrethandshake.EdKeyPair
	localID        string

	l net.Listener

	log   logging.Interface
	check = logging.CheckFatal

	verboseLogging bool

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

	ctx := context.Background()
	for {
		start := time.Now()
		conn, err := l.Accept()
		if err != nil {
			log.Log("error", "accept", "err", err)
			continue
		}

		go func() {
			rem := conn.RemoteAddr()
			shsAddr, ok := rem.(secretstream.Addr)
			if !ok {
				conn.Close()
				log.Log("err", errors.New("could not cast remote address"))
				return
			}
			id := SSBID(shsAddr.PubKey())
			log.Log("event", "connection established",
				"id", id,
				"addr", shsAddr.Addr.String(),
			)
			counter := debug.WrapCounter(conn)
			p := muxrpc.NewPacker(counter)
			if verboseLogging {
				p = muxrpc.NewPacker(codec.Wrap(kitlog.With(log, "id", id), counter))
			}
			handler := sbotHandler{id}
			rpc = muxrpc.Handle(p, handler)

			go serveRpc(ctx, start, id, rpc, counter)
		}()
	}
}
func serveRpc(ctx context.Context, start time.Time, id string, rpc muxrpc.Endpoint, counter *debug.Counter) {

	err := rpc.(muxrpc.Server).Serve(ctx)
	log.Log("event", "connection done",
		"id", id,
		"err", err,
		"took", time.Since(start),
		"sent", humanize.Bytes(counter.Cw.Count()),
		"rcvd", humanize.Bytes(counter.Cr.Count()),
	)
}

func initClient(ctx *cli.Context) error {
	verboseLogging = ctx.Bool("verbose")
	var err error
	localKey, err = secrethandshake.LoadSSBKeyPair(ctx.String("key"))
	if err != nil {
		return errors.Wrap(err, "failed to load keypair")
	}
	localID = SSBID(localKey.Public[:])
	srv, err := secretstream.NewServer(*localKey, sbotAppKey)
	if err != nil {
		return errors.Wrap(err, "failed to create server")
	}
	l, err = srv.Listen("tcp", ctx.String("addr"))
	if err != nil {
		return err
	}
	log.Log("msg", "listener created", "addr", l.Addr(), "ssbID", localID)

	return nil
}

func SSBID(pubKey []byte) string {
	return fmt.Sprintf("@%s.ed25519", base64.StdEncoding.EncodeToString(pubKey))
}

func serveCmd(ctx *cli.Context) error {
	/*
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
				if t, hasTimeout := args[0]["timeout"]; hasTimeout {
					timeout = t.(float64)
				}
				timeout = args[0].Timeout
			}
			log.Log("event", "incoming call", "cmd", "ping", "timeout", timeout)
			return struct {
				Pong string
			}{"test"}
		})
	*/

	/*
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
	*/

	return nil
}
