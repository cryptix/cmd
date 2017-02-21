// tnse stands for Test aNd Show Error
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/cryptix/go/logging"
	kitlog "github.com/go-kit/kit/log"
)

var log *kitlog.Context

// flags
var (
	timeout = flag.Duration("timeout", 3*time.Second, "how long to show the notifications")
)

func main() {
	logging.SetupLogging(nil)
	log = logging.Logger("tnse")

	flag.Parse()
	if len(os.Args) < 2 {
		logging.CheckFatal(fmt.Errorf("usage error"))
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
	var passed bool
	err = goTest.Run()
	done := time.Now()
	if err != nil {
		log = log.With("err", err)
		log.Log("event", "run failed")
	} else {
		passed = true
	}

	err = notify(passed, buf.String())
	logging.CheckFatal(err)
	log = log.With("took", done.Sub(start))
	log = log.With("passed", passed)
	log.Log("event", "'go test' finished")
}
