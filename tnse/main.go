// tnse stands for Test aNd Show Error
package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/cryptix/go/logging"
	"github.com/rs/xlog"
)

var log xlog.Logger

// flags
var (
	timeout = flag.Duration("timeout", 3*time.Second, "how long to show the notifications")
)

func main() {
	logging.SetupLogging(nil)
	log = logging.Logger("tnse")

	flag.Parse()
	if len(os.Args) < 2 {
		log.Fatal("usage error")
	}
	goTest := exec.Command("go", "test", os.Args[1])
	wd, err := os.Getwd()
	logging.CheckFatal(err)
	goTest.Dir = wd

	buf := new(bytes.Buffer)
	out := io.MultiWriter(buf, os.Stdout)
	goTest.Stderr = out
	goTest.Stdout = out

	start := time.Now()
	log.Debug("starting 'go test'")
	var passed bool
	err = goTest.Run()
	done := time.Now()
	if err != nil {
		log.SetField("err", err)
		log.Warn("run failed")
	} else {
		passed = true
	}

	err = notify(passed, buf.String())
	logging.CheckFatal(err)
	log.SetField("took", done.Sub(start))
	log.SetField("passed", passed)
	log.Info("'go test' finished")
}
