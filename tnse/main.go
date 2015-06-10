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
	"gopkg.in/errgo.v1"
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

	notify(passed, buf.String())
	log.WithFields(map[string]interface{}{
		"took":   done.Sub(start),
		"passed": passed,
	}).Infoln("'go test' finished")
}

func notify(passed bool, output string) error {
	lvl := "critical"
	title := "test failed"
	if passed {
		lvl = "normal"
		title = "test passed"
	}
	tout := fmt.Sprintf("%.0f", timeout.Seconds()*1000)
	xmsg := exec.Command("notify-send", "-t", tout, "-u", lvl, title, output)
	out, err := xmsg.CombinedOutput()
	if err != nil {
		return errgo.Notef(err, "notify-send failed: output: %s", out)
	}
	log.Debugln("notify-send:", out)
	return nil
}
