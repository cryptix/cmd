package main

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/cryptix/go-muxrpc"
	"github.com/cryptix/go-muxrpc/codec"
	"github.com/cryptix/secretstream"
	"github.com/cryptix/secretstream/secrethandshake"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/shurcooL/go-goon"
	"gopkg.in/urfave/cli.v2"
)

var sbotAppKey []byte
var defaultKeyFile string
var logger log.Logger

func init() {
	var err error
	sbotAppKey, err = base64.StdEncoding.DecodeString("1KHLiKZvAvjbY1ziZEHMXawbCEIM6qwjCDm3VYRan/s=")
	if err != nil {
		panic(err)
	}

	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	defaultKeyFile = filepath.Join(u.HomeDir, ".ssb", "secret")
}

var Revision = "unset"

func main() {

	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

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
			Name:   "hist",
			Action: createLogStream,
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
	if err := app.Run(os.Args); err != nil {
		logger.Log("error", err)
	}

}

var client *muxrpc.Client

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
		client = muxrpc.NewClient(logger, codec.Wrap(logger, conn))
	} else {
		client = muxrpc.NewClient(logger, conn)
	}
	return nil
}

func privatePublishCmd(ctx *cli.Context) error {
	content := map[string]interface{}{
		"text": ctx.String("text"),
		"type": ctx.String("type"),
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
	arg := map[string]interface{}{
		"content": content,
		"rcps":    recps,
	}
	var reply map[string]interface{}
	err := client.Call("private.publish", arg, &reply)
	if err != nil {
		return errors.Wrapf(err, "publish call failed.")
	}
	logger.Log("event", "private published")
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
	logger.Log("event", "get reply")
	goon.Dump(getReply)
	var reply map[string]interface{}
	if err := client.Call("private.unbox", getReply["content"], &reply); err != nil {
		return errors.Wrapf(err, "get call failed.")
	}
	logger.Log("event", "unboxed")
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
	logger.Log("event", "published")
	goon.Dump(reply)
	return client.Close()
}

func createHistoryStreamCmd(ctx *cli.Context) error {
	id := ctx.Args().Get(0)
	if id == "" {
		return errors.New("createHist: id can't be empty")
	}
	arg := map[string]interface{}{
		"content": id,
	}
	reply := make([]map[string]interface{}, 0, 10)

	err := client.SyncSource("createHistoryStream", arg, &reply)
	if err != nil {
		return errors.Wrapf(err, "createHistoryStream call failed.")
	}
	logger.Log("event", "got hist stream..!")
	goon.Dump(reply)
	return client.Close()
}

func createLogStream(ctx *cli.Context) error {
	reply := make([]map[string]interface{}, 0, 10)
	err := client.SyncSource("createLogStream", nil, &reply)
	if err != nil {
		return errors.Wrapf(err, "createLogStream call failed.")
	}
	logger.Log("event", "got log stream..!")
	for _, p := range reply {
		goon.Dump(p)
	}
	return client.Close()
}
func callCmd(ctx *cli.Context) error {
	cmd := ctx.Args().Get(0)
	if cmd == "" {
		return errors.New("call: cmd can't be empty")
	}
	arg := ctx.Args().Get(1)
	var reply interface{}
	if err := client.Call(cmd, arg, &reply); err != nil {
		return errors.Wrapf(err, "%s: call failed.", cmd)
	}
	logger.Log("event", "call reply")
	goon.Dump(reply)
	return client.Close()
}
