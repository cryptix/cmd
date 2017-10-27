package main

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/cryptix/go/logging"
	"github.com/cryptix/secretstream"
	"github.com/cryptix/secretstream/secrethandshake"
	"github.com/pkg/errors"
	"github.com/shurcooL/go-goon"
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
)

func init() {
	var err error
	sbotAppKey, err = base64.StdEncoding.DecodeString("1KHLiKZvAvjbY1ziZEHMXawbCEIM6qwjCDm3VYRan/s=")
	check(err)

	u, err := user.Current()
	check(err)

	defaultKeyFile = filepath.Join(u.HomeDir, ".ssb", "secret")
}

var Revision = "unset"

func main() {
	logging.SetupLogging(nil)
	log = logging.Logger("gophbot")

	app := cli.App{
		Name:    "ssb-gophbot",
		Usage:   "what can I say? sbot in Go",
		Version: "alpha2",
	}
	cli.VersionPrinter = func(c *cli.Context) {
		// go install -ldflags="-X main.Revision=$(git rev-parse HEAD)"
		fmt.Printf("%s ( rev: %s )\n", c.App.Version, Revision)
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "addr", Value: "localhost:8008", Usage: "tcp address of the sbot to connect to (or listen on)"},
		&cli.StringFlag{Name: "remoteKey", Value: "", Usage: "the remote pubkey you are connecting to (by default the local key)"},
		&cli.StringFlag{Name: "key,k", Value: defaultKeyFile},
		&cli.BoolFlag{Name: "verbose,vv", Usage: "print muxrpc packets"},
	}
	app.Before = initClient
	app.Commands = []*cli.Command{
		{
			Name:   "log",
			Action: logStreamCmd,
		},
		{
			Name:   "hist",
			Action: historyStreamCmd,
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "limit", Value: -1},
				&cli.IntFlag{Name: "seq", Value: 0},
				&cli.BoolFlag{Name: "reverse"},
				&cli.BoolFlag{Name: "live"},
				&cli.BoolFlag{Name: "keys", Value: true},
				&cli.BoolFlag{Name: "values", Value: true},
			},
		},
		{
			Name:   "qry",
			Action: query,
		},
		{
			Name:   "call",
			Action: callCmd,
			Usage:  "make an dump* async call",
			UsageText: `SUPPORTS:
* whoami
* latestSequence
* getLatest
* get
* blobs.(has|want|rm|wants)
* gossip.(peers|add|connect)


see https://scuttlebot.io/apis/scuttlebot/ssb.html#createlogstream-source  for more

CAVEAT: only one argument...
`,
		},
		{
			Name: "private",
			Subcommands: []*cli.Command{
				{
					Name:   "publish",
					Usage:  "p",
					Action: privatePublishCmd,
					Flags: []cli.Flag{
						&cli.StringFlag{Name: "type", Value: "post"},
						&cli.StringFlag{Name: "text", Value: "Hello, World!"},
						&cli.StringFlag{Name: "root", Usage: "the ID of the first message of the thread"},
						&cli.StringFlag{Name: "branch", Usage: "the post ID that is beeing replied to"},
						&cli.StringFlag{Name: "channel"},
						&cli.StringSliceFlag{Name: "recps", Usage: "posting to these IDs privatly"},
					},
				},
				{
					Name:   "unbox",
					Usage:  "u",
					Action: privateUnboxCmd,
				},
			},
		},
		{
			Name:   "publish",
			Usage:  "p",
			Action: publishCmd,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "type", Value: "post"},
				&cli.StringFlag{Name: "text", Value: "Hello, World!"},
				&cli.StringFlag{Name: "root", Value: "", Usage: "the ID of the first message of the thread"},
				&cli.StringFlag{Name: "branch", Value: "", Usage: "the post ID that is beeing replied to"},
			},
		},
	}
	check(app.Run(os.Args))
}

func initClient(ctx *cli.Context) error {
	localKey, err := secrethandshake.LoadSSBKeyPair(ctx.String("key"))
	if err != nil {
		return err
	}
	var conn net.Conn
	if ctx.Bool("listen") { // TODO: detect server command..
		srv, err := secretstream.NewServer(*localKey, sbotAppKey)
		if err != nil {
			return err
		}
		l, err := srv.Listen("tcp", ctx.String("addr"))
		if err != nil {
			return err
		}
		conn, err = l.Accept()
		if err != nil {
			return err
		}
	} else {
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
	return nil
}

func privatePublishCmd(ctx *cli.Context) error {
	content := map[string]interface{}{
		"text": ctx.String("text"),
		"type": ctx.String("type"),
	}
	if c := ctx.String("channel"); c != "" {
		content["channel"] = c
	}
	if r := ctx.String("root"); r != "" {
		content["root"] = r
		if b := ctx.String("branch"); b != "" {
			content["branch"] = b
		} else {
			content["branch"] = r
		}
	}
	recps := ctx.StringSlice("recps")
	if len(recps) == 0 {
		return errors.Errorf("private.publish: 0 recps.. that would be quite the lonely message..")
	}
	var reply map[string]interface{}
	err := client.Call("private.publish", &reply, content, recps)
	if err != nil {
		return errors.Wrapf(err, "publish call failed.")
	}
	log.Log("event", "private published")
	goon.Dump(reply)
	return client.Close()
}

func privateUnboxCmd(ctx *cli.Context) error {
	id := ctx.Args().Get(0)
	if id == "" {
		return errors.New("get: id can't be empty")
	}
	var getReply map[string]interface{}
	if err := client.Call("get", id, &getReply); err != nil {
		return errors.Wrapf(err, "get call failed.")
	}
	log.Log("event", "get reply")
	goon.Dump(getReply)
	var reply map[string]interface{}
	if err := client.Call("private.unbox", getReply["content"], &reply); err != nil {
		return errors.Wrapf(err, "get call failed.")
	}
	log.Log("event", "unboxed")
	goon.Dump(reply)
	return client.Close()
}

func publishCmd(ctx *cli.Context) error {
	arg := map[string]interface{}{
		"text": ctx.String("text"),
		"type": ctx.String("type"),
	}
	if r := ctx.String("root"); r != "" {
		arg["root"] = r
		if b := ctx.String("branch"); b != "" {
			arg["branch"] = b
		} else {
			arg["branch"] = r
		}
	}
	var reply map[string]interface{}
	err := client.Call("publish", arg, &reply)
	if err != nil {
		return errors.Wrapf(err, "publish call failed.")
	}
	log.Log("event", "published")
	goon.Dump(reply)
	return client.Close()
}

func historyStreamCmd(ctx *cli.Context) error {
	id := ctx.Args().Get(0)
	if id == "" {
		return errors.New("createHist: id can't be empty")
	}
	arg := map[string]interface{}{
		"id":      id,
		"limit":   ctx.Int("limit"),
		"seq":     ctx.Int("seq"),
		"live":    ctx.Bool("live"),
		"reverse": ctx.Bool("reverse"),
		"keys":    ctx.Bool("keys"),
		"values":  ctx.Bool("values"),
	}
	reply := make(chan map[string]interface{})
	go func() {
		for r := range reply {
			goon.Dump(r)
		}
	}()
	if err := client.Source("createHistoryStream", reply, arg); err != nil {
		return errors.Wrap(err, "source stream call failed")
	}
	return client.Close()
}

func logStreamCmd(ctx *cli.Context) error {
	reply := make(chan map[string]interface{})
	go func() {
		for r := range reply {
			goon.Dump(r)
		}
	}()
	if err := client.Source("createLogStream", reply); err != nil {
		return errors.Wrap(err, "source stream call failed")
	}
	return client.Close()
}

func callCmd(ctx *cli.Context) error {
	cmd := ctx.Args().Get(0)
	if cmd == "" {
		return errors.New("call: cmd can't be empty")
	}
	var reply interface{}
	if err := client.Call(cmd, &reply, ctx.Args().Slice()); err != nil {
		return errors.Wrapf(err, "%s: call failed.", cmd)
	}
	log.Log("event", "call reply")
	goon.Dump(reply)
	return client.Close()
}

func query(ctx *cli.Context) error {
	reply := make(chan map[string]interface{})
	go func() {
		for r := range reply {
			goon.Dump(r)
		}
	}()
	if err := client.Source("query.read", reply, ctx.Args().Get(0)); err != nil {
		return errors.Wrap(err, "source stream call failed")
	}
	return client.Close()
}
