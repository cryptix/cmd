package main

import (
	"fmt"
	"os/exec"

	"gopkg.in/errgo.v1"
)

func notify(passed bool, output string) error {
	lvl := "critical"
	title := "test failed"
	if passed {
		lvl = "normal"
		title = "test passed"
	}
	tout := fmt.Sprintf("%.0f", timeout.Seconds()*1000)
	if len(output) > 300 {
		output = output[:300]
	}
	output = `'` + output + `'`
	xmsg := exec.Command("notify-send", "-t", tout, "-u", lvl, title, output)
	out, err := xmsg.CombinedOutput()
	if err != nil {
		return errgo.Notef(err, "notify-send failed: output: %s", out)
	}
	return nil
}
