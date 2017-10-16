package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cryptoscope.co/go/binpath"

	"cryptoscope.co/go/panopticon"
	"cryptoscope.co/go/voyeur"
	"github.com/cryptix/go/logging"
	"github.com/cryptix/secretstream"
	"github.com/cryptix/secretstream/secrethandshake"
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v2"
	"scuttlebot.io/go/muxrpc"
	"scuttlebot.io/go/muxrpc/codec"
)

var (
	sbotAppKey     []byte
	defaultKeyFile string
	client         *muxrpc.Client

	log   logging.Interface
	check = logging.CheckFatal

	bdb    *badger.DB
	pstore *panopticon.Store
)

func init() {
	var err error
	sbotAppKey, err = base64.StdEncoding.DecodeString("1KHLiKZvAvjbY1ziZEHMXawbCEIM6qwjCDm3VYRan/s=")
	check(err)

	u, err := user.Current()
	check(err)

	defaultKeyFile = filepath.Join(u.HomeDir, ".ssb", "secret")
}

func initClient(ctx *cli.Context) error {
	localKey, err := secrethandshake.LoadSSBKeyPair(ctx.String("key"))
	if err != nil {
		return err
	}
	var conn net.Conn
	c, err := secretstream.NewClient(*localKey, sbotAppKey)
	if err != nil {
		return err
	}
	var remotPubKey = localKey.Public
	if rk := ctx.String("remoteKey"); rk != "" {
		rk = strings.TrimSuffix(rk, ".ed25519")
		rk = strings.TrimPrefix(rk, "@")
		rpk, err := base64.StdEncoding.DecodeString(rk)
		if err != nil {
			return errors.Wrapf(err, "ssb-gophbot: base64 decode of --remoteKey failed")
		}
		copy(remotPubKey[:], rpk)
	}
	d, err := c.NewDialer(remotPubKey)
	if err != nil {
		return err
	}
	conn, err = d("tcp", ctx.String("addr"))
	if err != nil {
		return err
	}
	if ctx.Bool("verbose") {
		client = muxrpc.NewClient(log, codec.Wrap(log, conn))
	} else {
		client = muxrpc.NewClient(log, conn)
	}
	go func() {
		client.Handle()
		log.Log("warning", "muxrpc disconnected")
	}()

	opts := badger.DefaultOptions
	opts.Dir = ctx.String("db")
	opts.ValueDir = ctx.String("db")
	bdb, err = badger.Open(opts)
	check(err)
	pstore = panopticon.NewStore(bdb, voyeur.Fwd)
	return nil
}

var Revision = "unset"

func main() {
	logging.SetupLogging(nil)
	log = logging.Logger("gophermit")

	app := cli.App{
		Name:    "ssb-gophermit",
		Usage:   "a panoptical hermit in go",
		Version: "alpha1",
	}
	cli.VersionPrinter = func(c *cli.Context) {
		// go install -ldflags="-X main.Revision=$(git rev-parse HEAD)"
		fmt.Printf("%s ( rev: %s )\n", c.App.Version, Revision)
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "db", Value: "./db"},
		&cli.StringFlag{Name: "addr", Value: "localhost:8008", Usage: "tcp address of the sbot to connect to (or listen on)"},
		&cli.StringFlag{Name: "remoteKey", Value: "", Usage: "the remote pubkey you are connecting to (by default the local key)"},
		&cli.StringFlag{Name: "key,k", Value: defaultKeyFile},
		&cli.BoolFlag{Name: "verbose,vv", Usage: "print muxrpc packets"},
	}
	app.Before = initClient
	app.Commands = []*cli.Command{
		{
			Name:   "ls",
			Action: lsCmd,
		},
		{
			Name:   "slurp",
			Action: slurpCmd,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "id", Value: "@p13zSAiOpguI9nsawkGijsnMfWmFd5rlUNpzekEE+vI=.ed25519"},
				&cli.IntFlag{Name: "limit", Value: -1},
				&cli.IntFlag{Name: "seq", Value: 0},
			},
		},
		{
			Name:   "purge",
			Action: purgeCmd,
		},
	}

	check(app.Run(os.Args))
}

func purgeCmd(ctx *cli.Context) error {
	check(bdb.PurgeOlderVersions())
	r, err := strconv.ParseFloat(ctx.Args().Get(0), 10)
	check(err)
	check(bdb.RunValueLogGC(r))
	return bdb.Close()
}

func lsCmd(ctx *cli.Context) error {
	opt := badger.DefaultIteratorOptions
	opt.PrefetchSize = 50
	return bdb.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(opt)
		for it.Rewind(); it.Valid(); it.Next() {
			i := it.Item()
			fmt.Printf("k: %q %d\n", i.Key(), i.Version())

			/*
				v, err := i.Value()
				check(err)
				var msg map[string]interface{}
				err = json.NewDecoder(bytes.NewReader(v)).Decode(&msg)
				check(err)
				goon.Dump(msg)
			*/
		}
		return nil
	})
}

type testEvent struct{ Msg map[string]interface{} }

func (te testEvent) EventType() string { return "testev" }

var msgpath = binpath.Must(binpath.FromString("msgs"))

func slurpRouter(ctx context.Context, s *panopticon.Store, em voyeur.Emitter, ev voyeur.Event) {
	msg := ev.(testEvent).Msg

	mk, err := binpath.FromString(msg["key"].(string))
	check(err)

	var b bytes.Buffer
	check(json.NewEncoder(&b).Encode(msg["value"]))

	err = s.Put(binpath.Join(msgpath, mk), b.Bytes())
	check(err)
}

func slurpCmd(ctx *cli.Context) error {
	_, err := pstore.MkSubStore("route", slurpRouter)
	if err != nil {
		return errors.Wrap(err, "could not MkSubStore")
	}
	msgs := make(chan map[string]interface{})
	go func() {
		i := 1
		start := time.Now()
		for r := range msgs {
			pstore.OnEvent(context.TODO(), testEvent{r})
			if i%1000 == 0 {
				log.Log("msg", "processed", "i", i, "took", fmt.Sprintf("%v", time.Since(start)))
				start = time.Now()
			}
			i++
		}
	}()
	opts := map[string]interface{}{
		"id":    ctx.String("id"),
		"limit": ctx.Int("limit"),
		"seq":   ctx.Int("seq"),
	}
	if err := client.Source("createHistoryStream", msgs, opts); err != nil {
		log.Log("err", errors.Wrap(err, "source stream call failed"))
	}
	check(bdb.Close())
	return client.Close()
}
