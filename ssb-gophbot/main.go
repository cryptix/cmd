package main

import (
	"encoding/base64"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/cryptix/go-muxrpc"
	"github.com/cryptix/go-muxrpc/codec"
	"github.com/cryptix/secretstream"
	"github.com/cryptix/secretstream/secrethandshake"
	"github.com/shurcooL/go-goon"
	"gopkg.in/errgo.v1"
)

var sbotAppKey []byte
var defaultKeyFile string

func init() {
	var err error
	sbotAppKey, err = base64.StdEncoding.DecodeString("1KHLiKZvAvjbY1ziZEHMXawbCEIM6qwjCDm3VYRan/s=")
	if err != nil {
		log.Fatal(err)
	}

	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	defaultKeyFile = filepath.Join(u.HomeDir, ".ssb", "secret")
}

func main() {
	app := cli.NewApp()
	app.Name = "ssb-gophbot"
	app.Usage = "what can I say? sbot in Go"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "addr", Value: "localhost:8008", Usage: "tcp address of the sbot to connect to (or listen on)"},
		cli.StringFlag{Name: "remoteKey", Value: "", Usage: "the remote pubkey you are connecting to (by default the local key)"},
		cli.StringFlag{Name: "key,k", Value: defaultKeyFile},
		cli.BoolFlag{Name: "verbose,vv", Usage: "print muxrpc packets"},
	}
	app.Before = initClient
	app.Commands = []cli.Command{
		{
			Name:   "whoami",
			Action: whoamiCmd,
		},
		{
			Name:   "publish",
			Usage:  "p",
			Action: publishCmd,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "type", Value: "post"},
				cli.StringFlag{Name: "text", Value: "Hello, World!"},
				cli.StringFlag{Name: "root", Value: "", Usage: "the ID of the first message of the thread"},
				cli.StringFlag{Name: "branch", Value: "", Usage: "the post ID that is beeing replied to"},
			},
		},
	}
	app.Run(os.Args)
}

var client *muxrpc.Client

func initClient(ctx *cli.Context) error {
	log.SetOutput(os.Stderr)

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

		d, err := c.NewDialer(localKey.Public)
		if err != nil {
			return err
		}

		conn, err = d("tcp", ctx.String("addr"))
		if err != nil {
			return err
		}
	}

	if ctx.Bool("verbose") {
		client = muxrpc.NewClient(codec.Wrap(conn))
	} else {
		client = muxrpc.NewClient(conn)
	}
	return nil
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
		return errgo.Notef(err, "publish call failed.")
	}
	log.Println("published..!")
	goon.Dump(reply)
	return client.Close()
}

func createHistoryStreamCmd(ctx *cli.Context) error {
	id := ctx.Args().Get(0)
	if id == "" {
		return errgo.New("post: id can't be empty")
	}
	arg := map[string]interface{}{
		"content": id,
	}
	reply := make([]map[string]interface{}, 0, 10)
	err := client.SyncSource("createHistoryStream", arg, &reply)
	if err != nil {
		return errgo.Notef(err, "createHistoryStream call failed.")
	}
	log.Println("got hist stream..!")
	goon.Dump(reply)
	return client.Close()
}

func whoamiCmd(ctx *cli.Context) error {

	// go func() {
	// 	reply := make([]map[string]interface{}, 0, 10)
	// 	err := c.SyncSource("createLogStream", nil, &reply)
	// 	check(err)
	// 	log.Println("got log stream..!")
	// 	for _, p := range reply {
	// 		goon.Dump(p)
	// 	}
	// }()

	var reply map[string]interface{}
	if err := client.Call("whoami", nil, &reply); err != nil {
		return errgo.Notef(err, "whoami call failed.")
	}
	// goon.Dump(reply)
	log.Print("ID:", reply["id"])
	return client.Close()
}
