package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cryptoscope.co/go/binpath"
	"cryptoscope.co/go/specialκ"
	"cryptoscope.co/go/specialκ/persistent"
	"github.com/cryptix/go/logging"
	"github.com/cryptix/secretstream"
	"github.com/cryptix/secretstream/secrethandshake"
	"github.com/dgraph-io/badger"
	kitlog "github.com/go-kit/kit/log"
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
	pstore specialκ.MFR

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
	pstore = persistent.New(persistent.JSONCodec, bdb, log)
	return nil
}

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
	pref, err := binpath.ParseString(ctx.Args().First())
	if err != nil {
		return err
	}
	opt := badger.DefaultIteratorOptions
	opt.PrefetchSize = 50
	return bdb.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opt)

		for it.Seek(pref); it.ValidForPrefix(pref); it.Next() {
			i := it.Item()
			v, err := i.Value()
			if err != nil {
				return err
			}
			fmt.Printf("%s: %d %x\n", binpath.Path(i.Key()), len(v), i.UserMeta())
		}
		return nil
	})
}

type ssbMsg map[string]interface{}

func slurpCmd(c *cli.Context) error {
	start := time.Now()

	emitter, src := pstore.Pair(ssbMsg{})

	ctx := context.TODO()
	specialκ.Then(ctx, src, map[string]specialκ.Sink{
		"author": pstore.Map(ssbMsg{}, func(_ context.Context, e specialκ.Entry) specialκ.Entry {
			msg, ok := e.Value.(ssbMsg)
			if ok {
				author := c2m(msg, "value")["author"].(string)
				seq := c2m(msg, "value")["sequence"].(float64)
				e.Prefix = binpath.JoinStrings("author", author)
				e.Key = binpath.FromUint64(uint64(seq))
				e.Value = e.Seq
			}
			return e
		}),
		"type": pstore.Map(ssbMsg{}, func(_ context.Context, e specialκ.Entry) specialκ.Entry {
			msg, ok := e.Value.(ssbMsg)
			if ok {
				content := c2m(msg, "value", "content")
				var t string
				if content == nil {
					t = "string"
				} else {
					t = content["type"].(string)
				}
				e.Prefix = binpath.JoinStrings("type", t)
				e.Key = binpath.FromString(msg["key"].(string))
				// TODO: reduce to ssb-host struct
			}
			return e
		}),
		"pub": pstore.Filter(ssbMsg{}, func(_ context.Context, e specialκ.Entry) bool {
			msg, ok := e.Value.(ssbMsg)
			if ok {
				content := c2m(msg, "value", "content")
				if content == nil {
					return false
				}
				if t := content["type"].(string); t == "pub" {
					return true
				}
			}
			return false
		}),
	}, kitlog.NewNopLogger())

	var i uint64
	msgs := make(chan ssbMsg)
	wait := make(chan bool)
	go func() {
		last := time.Now()
		for r := range msgs {
			emitter.Emit(ctx, specialκ.Entry{
				Seq:   i,
				Value: r,
				Key:   binpath.FromString(r["key"].(string)),
			})
			i++
			if i%1000 == 0 {
				log.Log("msg", "processed", "i", i, "took", fmt.Sprintf("%v", time.Since(last)))
				last = time.Now()
			}
		}
		wait <- true
	}()
	opts := map[string]interface{}{
		"id":    c.String("id"),
		"limit": c.Int("limit"),
		"seq":   c.Int("seq"),
	}
	if err := client.Source("createHistoryStream", msgs, opts); err != nil {
		log.Log("warning", errors.Wrap(err, "source stream call failed"))
	}
	close(msgs)
	log.Log("done", "slurp", "msgs", i-1, "id", c.String("id"), "took", fmt.Sprintf("%v", time.Since(start)))
	<-wait
	check(bdb.Close())
	return client.Close()
}

func c2m(v map[string]interface{}, fields ...string) map[string]interface{} {
	var ok bool
	for _, f := range fields {
		v, ok = v[f].(map[string]interface{})
		if !ok {
			return nil
		}
	}
	return v
}
