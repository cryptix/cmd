// tnse stands for Test aNd Show Error
package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"time"

	"gopkg.in/errgo.v1"

	"github.com/cryptix/go/logging"
)

var log = logging.Logger("tnse")

func main() {
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
		notify(passed, buf.String())
	} else {
		passed = true
		notify(passed, "")
	}
	log.WithFields(map[string]interface{}{
		"took":   done.Sub(start),
		"passed": passed,
	}).Infoln("'go test' finished")
}

func notify(passed bool, output string) error {
	lvl := "critical"
	title := "passed"
	if passed {
		lvl = "normal"
		title = "passed"
	}
	xmsg := exec.Command("notify-send", "-t", "3000", "-u", lvl, "test "+title, output)
	out, err := xmsg.CombinedOutput()
	if err != nil {
		return errgo.Notef(err, "notify-send failed: output: %s", out)
	}
	log.Debugln("notify-send:", out)
	return nil
}
