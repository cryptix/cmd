package main

import (
	"os/exec"

	"gopkg.in/errgo.v1"
)

// uses https://github.com/alloy/terminal-notifier

func notify(passed bool, output string) error {
	subTitle := "🚨test failed🚨"
	if passed {
		subTitle = "🌟test passed🌟"
	}

	xmsg := exec.Command("open", "-a", "terminal-notifier.app", "--args",
		"-title", "testNshowErr",
		"-subtitle", subTitle, "-message", output)
	out, err := xmsg.CombinedOutput()
	if err != nil {
		return errgo.Notef(err, "terminal-notifier.app failed: output: %s", out)
	}
	log.Debugln("terminal-notifier.app:", out)
	return nil
}
