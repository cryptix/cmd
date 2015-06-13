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
)

var log = logging.Logger("tnse")

// flags
var (
	timeout = flag.Duration("timeout", 3*time.Second, "how long to show the notifications")
)

func main() {
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
	log.Debugln("starting 'go test'")
	var passed bool
	err = goTest.Run()
	done := time.Now()
	if err != nil {
		log.WithField("err", err.Error()).Warning("run failed")
	} else {
		passed = true
	}

	err = notify(passed, buf.String())
	logging.CheckFatal(err)
	log.WithFields(map[string]interface{}{
		"took":   done.Sub(start),
		"passed": passed,
	}).Infoln("'go test' finished")
}
