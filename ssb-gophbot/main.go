package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"

	"github.com/cryptix/go-muxrpc"
	"github.com/cryptix/go-muxrpc/codec"
	"github.com/cryptix/secretstream"
	"github.com/cryptix/secretstream/secrethandshake"
	"github.com/shurcooL/go-goon"
	"gopkg.in/errgo.v1"
	"gopkg.in/urfave/cli.v2"
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

var Revision = "unset"

func main() {
	app := cli.NewApp()
	app.Name = "ssb-gophbot"
	app.Usage = "what can I say? sbot in Go"

	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t" }}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
VERSION:
   {{.Version}}
`

	cli.CommandHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} [global options] {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}

{{.UsageText}}
`

	cli.VersionPrinter = func(c *cli.Context) {
		// go install -ldflags="-X main.Revision=$(git rev-parse HEAD)"
		fmt.Printf("%s ( rev: %s )\n", c.App.Version, Revision)
	}
	app.Version = "alpha1"

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
		log.Println("Error: ", err)
	}

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
		return errgo.Newf("private.publish: 0 recps.. that would be quite the lonely message..")
	}
	arg := map[string]interface{}{
		"content": content,
		"rcps":    recps,
	}
	var reply map[string]interface{}
	err := client.Call("private.publish", arg, &reply)
	if err != nil {
		return errgo.Notef(err, "publish call failed.")
	}
	log.Println("private published..!")
	goon.Dump(reply)
	return client.Close()
}

func privateUnboxCmd(ctx *cli.Context) error {
	id := ctx.Args().Get(0)
	if id == "" {
		return errgo.New("get: id can't be empty")
	}
	var getReply map[string]interface{}
	if err := client.Call("get", id, &getReply); err != nil {
		return errgo.Notef(err, "get call failed.")
	}
	log.Print("get:")
	goon.Dump(getReply)

	var reply map[string]interface{}
	if err := client.Call("private.unbox", getReply["content"], &reply); err != nil {
		return errgo.Notef(err, "get call failed.")
	}

	log.Print("unbox:")
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
		return errgo.Notef(err, "publish call failed.")
	}
	log.Println("published..!")
	goon.Dump(reply)
	return client.Close()
}

func createHistoryStreamCmd(ctx *cli.Context) error {
	id := ctx.Args().Get(0)
	if id == "" {
		return errgo.New("createHist: id can't be empty")
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

func createLogStream(ctx *cli.Context) error {
	reply := make([]map[string]interface{}, 0, 10)
	err := client.SyncSource("createLogStream", nil, &reply)
	if err != nil {
		return errgo.Notef(err, "createLogStream call failed.")
	}
	log.Println("got log stream..!")
	for _, p := range reply {
		goon.Dump(p)
	}
	return client.Close()
}
func callCmd(ctx *cli.Context) error {
	cmd := ctx.Args().Get(0)
	if cmd == "" {
		return errgo.New("call: cmd can't be empty")
	}
	arg := ctx.Args().Get(1)
	var reply interface{}
	if err := client.Call(cmd, arg, &reply); err != nil {
		return errgo.Notef(err, "%s: call failed.", cmd)
	}
	log.Print("call:")
	goon.Dump(reply)
	return client.Close()
}
