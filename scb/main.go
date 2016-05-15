package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/cryptix/go-muxrpc"
	"github.com/cryptix/secretstream"
	"github.com/cryptix/secretstream/secrethandshake"
	"github.com/shurcooL/go-goon"
)

var sbotAppKey []byte
var defaultKeyFile string

func init() {
	var err error
	sbotAppKey, err = base64.StdEncoding.DecodeString("1KHLiKZvAvjbY1ziZEHMXawbCEIM6qwjCDm3VYRan/s=")
	check(err)

	u, err := user.Current()
	check(err)

	defaultKeyFile = filepath.Join(u.HomeDir, ".ssb", "secret")

}

func main() {
	app := cli.NewApp()
	app.Name = "scb"
	app.Usage = "securly boxes your cats from a to b"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "listen,l"},
		cli.StringFlag{Name: "key,k", Value: defaultKeyFile},
		cli.StringFlag{Name: "port,p", Value: "undefined", Usage: "the namespace to use"},
		cli.DurationFlag{Name: "timeout,t", Value: time.Minute},
		cli.BoolFlag{Name: "verbose,vv", Usage: "print gathered stats to stderr"},
	}
	app.Run(os.Args)
}

func run(ctx *cli.Context) {
	log.SetOutput(os.Stderr)
	if ctx.Bool("verbose") {
	}

	localKey := mustLoadKeyPair(ctx.String("key"))

	var conn net.Conn
	if ctx.Bool("listen") {
		srv, err := secretstream.NewServer(localKey, sbotAppKey)
		check(err)

		l, err := srv.Listen("tcp", ctx.Args().Get(0))
		check(err)

		conn, err = l.Accept()
		check(err)
	} else {
		var remotepub [32]byte
		rp, err := base64.StdEncoding.DecodeString(strings.TrimSuffix(ctx.Args().Get(1), ".ed25519"))
		check(err)
		copy(remotepub[:], rp)

		c, err := secretstream.NewClient(localKey, sbotAppKey)
		check(err)

		d, err := c.NewDialer(remotepub)
		check(err)

		conn, err = d("tcp", ctx.Args().Get(0))
		check(err)
	}

	beepBoop(conn)
}

func beepBoop(conn net.Conn) {
	// c := muxrpc.NewClient(codec.Wrap(conn))
	c := muxrpc.NewClient(conn)

	// go func() {
	// 	reply := make([]map[string]interface{}, 0, 10)
	// 	err := c.SyncSource("createLogStream", nil, &reply)
	// 	check(err)
	// 	log.Println("got log stream..!")
	// 	for _, p := range reply {
	// 		goon.Dump(p)
	// 	}
	// }()

	// go func() {
	// 	arg := map[string]interface{}{
	// 		"id": "@p13zSAiOpguI9nsawkGijsnMfWmFd5rlUNpzekEE+vI=.ed25519",
	// 	}
	// 	reply := make([]map[string]interface{}, 0, 10)
	// 	err := c.SyncSource("createHistoryStream", arg, &reply)
	// 	check(err)
	// 	log.Println("got hist stream..!")
	// 	goon.Dump(reply)
	// }()

	// go func() {
	// 	for {
	// 		log.Println("who am i..?")
	// 		var reply map[string]interface{}
	// 		if err := c.Call("whoami", nil, &reply); err != nil {
	// 			log.Println("no whoami")
	// 			break
	// 		}
	// 		goon.Dump(reply)
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

	for {
		log.Println("where am i..?")
		time.Sleep(1 * time.Second)
	}

	// echo!
	//_, err := io.Copy(conn, conn)
	//check(err)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mustLoadKeyPair(fname string) secrethandshake.EdKeyPair {
	f, err := os.Open(fname)
	check(err)

	var sbotKey struct {
		Curve   string `json:"curve"`
		ID      string `json:"id"`
		Private string `json:"private"`
		Public  string `json:"public"`
	}

	check(json.NewDecoder(f).Decode(&sbotKey))

	public, err := base64.StdEncoding.DecodeString(strings.TrimSuffix(sbotKey.Public, ".ed25519"))
	check(err)

	private, err := base64.StdEncoding.DecodeString(strings.TrimSuffix(sbotKey.Private, ".ed25519"))
	check(err)

	var kp secrethandshake.EdKeyPair
	copy(kp.Public[:], public)
	copy(kp.Secret[:], private)
	return kp
}
