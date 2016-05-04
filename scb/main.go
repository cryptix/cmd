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
	"github.com/cryptix/go-shs"
	"github.com/keks/boxstream"
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

	if ctx.Bool("listen") {
		l, err := net.Listen("tcp", ctx.Args().Get(0))
		check(err)
		conn, err := l.Accept()
		check(err)

		servState, err := shs.NewServerState(sbotAppKey, localKey)
		check(err)

		err = shs.Server(servState, conn)
		check(err)

		en_k, en_n := servState.GetBoxstreamEncKeys()
		conn_w := boxstream.NewBoxer(conn, &en_n, &en_k)

		de_k, de_n := servState.GetBoxstreamDecKeys()
		conn_r := boxstream.NewUnboxer(conn, &de_n, &de_k)

		boxed := Conn{conn_r, conn_w, conn}

		beepBoop(boxed)
		return
	}

	var remotepub [32]byte
	rp, err := base64.StdEncoding.DecodeString(strings.TrimSuffix(ctx.Args().Get(1), ".ed25519"))
	check(err)
	copy(remotepub[:], rp)

	conn, err := net.Dial("tcp", ctx.Args().Get(0))
	check(err)

	state, err := shs.NewClientState(sbotAppKey, localKey, remotepub)
	check(err)

	check(shs.Client(state, conn))

	en_k, en_n := state.GetBoxstreamEncKeys()
	conn_w := boxstream.NewBoxer(conn, &en_n, &en_k)

	de_k, de_n := state.GetBoxstreamDecKeys()
	conn_r := boxstream.NewUnboxer(conn, &de_n, &de_k)

	boxed := Conn{conn_r, conn_w, conn}

	beepBoop(boxed)
}

func beepBoop(conn net.Conn) {
	c := muxrpc.NewClient(conn)
	/*
		go func() {
			reply := make([]map[string]interface{}, 0, 10)
			err := c.SyncSource("createLogStream", nil, &reply)
			check(err)
			log.Println("got log stream..!")
		}()
		go func() {
			for {
				log.Println("who am i..?")
				var reply interface{}
				if err := c.Call("whoami", nil, &reply); err != nil {
					log.Println("no whoami")
					break
				}
				time.Sleep(1 * time.Second)
			}
		}()
	*/
	go func() {
		arg := map[string]interface{}{
			"id": "@p13zSAiOpguI9nsawkGijsnMfWmFd5rlUNpzekEE+vI=.ed25519",
		}
		reply := make([]map[string]interface{}, 0, 10)
		err := c.SyncSource("createHistoryStream", arg, &reply)
		check(err)
		log.Println("got hist stream..!")
		goon.Dump(reply)
	}()
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

func mustLoadKeyPair(fname string) shs.EdKeyPair {
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

	var kp shs.EdKeyPair
	copy(kp.Public[:], public)
	copy(kp.Secret[:], private)
	return kp
}
