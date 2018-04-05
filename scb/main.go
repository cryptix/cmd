package main // import "github.com/cryptix/cmd/scb"

import (
	"encoding/base64"
	"errors"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"cryptoscope.co/go/secretstream"
	"cryptoscope.co/go/secretstream/secrethandshake"
	"github.com/codegangsta/cli"
	"github.com/cryptix/go/logging"
	humanize "github.com/dustin/go-humanize"
	kitlog "github.com/go-kit/kit/log"
	"github.com/miolini/datacounter"
)

var (
	sbotAppKey     []byte
	defaultKeyFile string
	log            kitlog.Logger
)

func init() {
	u, err := user.Current()
	logging.CheckFatal(err)

	defaultKeyFile = filepath.Join(u.HomeDir, ".ssb", "secret")
}

func main() {
	app := cli.NewApp()
	app.Name = "scb"
	app.Usage = "securly boxes your cats from a to b"
	app.Version = "0.1"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "listen,l"},
		cli.StringFlag{Name: "appKey", Value: "sAI1D9LtA6TQj3qj59/3bdKqiv4QQC3DY6/fWzBTDr8=", Usage: "the shared secret/mac key (in base64, plz)"},
		cli.StringFlag{Name: "key,k", Value: defaultKeyFile, Usage: "the ssb keyfile to load the local keypair from"},
	}
	logging.SetupLogging(nil)
	log = logging.Logger(app.Name)
	app.Run(os.Args)
}

func run(ctx *cli.Context) error {

	localKey, err := secrethandshake.LoadSSBKeyPair(ctx.String("key"))
	logging.CheckFatal(err)

	sbotAppKey, err = base64.StdEncoding.DecodeString(ctx.String("appKey"))
	logging.CheckFatal(err)

	var conn net.Conn
	if ctx.Bool("listen") {
		srv, err := secretstream.NewServer(*localKey, sbotAppKey)
		logging.CheckFatal(err)

		l, err := srv.Listen("tcp", ctx.Args().Get(0))
		logging.CheckFatal(err)

		log.Log("event", "listening",
			"id", base64.StdEncoding.EncodeToString(localKey.Public[:]),
			"addr", l.Addr().String(),
		)
		conn, err = l.Accept()
		logging.CheckFatal(err)

	} else {
		var remotepub [32]byte
		rp, err := base64.StdEncoding.DecodeString(strings.TrimSuffix(ctx.Args().Get(1), ".ed25519"))
		logging.CheckFatal(err)
		copy(remotepub[:], rp)

		c, err := secretstream.NewClient(*localKey, sbotAppKey)
		logging.CheckFatal(err)

		d, err := c.NewDialer(remotepub)
		logging.CheckFatal(err)

		conn, err = d("tcp", ctx.Args().Get(0))
		logging.CheckFatal(err)
	}
	start := time.Now()

	// showing off a little...
	rem := conn.RemoteAddr()
	shsAddr, ok := rem.(secretstream.Addr)
	if !ok {
		logging.CheckFatal(errors.New("could not cast remote address"))
	}
	log.Log("event", "connection established",
		"id", base64.StdEncoding.EncodeToString(shsAddr.PubKey()),
		"addr", shsAddr.Addr.String(),
	)

	var sentCounter, recvdCounter *datacounter.ReaderCounter
	go func() {
		sentCounter = datacounter.NewReaderCounter(os.Stdin)
		_, err := io.Copy(conn, sentCounter)
		logging.CheckFatal(err)
		conn.Close()
	}()

	recvdCounter = datacounter.NewReaderCounter(conn)
	_, err = io.Copy(os.Stdout, recvdCounter)

	log.Log("event", "copy done",
		"took", time.Since(start),
		"sent", humanize.Bytes(sentCounter.Count()),
		"rcvd", humanize.Bytes(recvdCounter.Count()),
	)
	return err
}
