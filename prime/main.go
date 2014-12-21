// prime reruns a command until it finishes under a certain limit.
// usefull for filling caches.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/cryptix/go/logging"
)

var limit = flag.Duration("d", 10*time.Second, "what is the limit to stop at")

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatalln("Not enough arguments")
	}

	for {

		start := time.Now()
		cmd := exec.Command(args[0], args[1:]...)

		stdout, err := cmd.StdoutPipe()
		logging.CheckFatal(err)

		stderr, err := cmd.StderrPipe()
		logging.CheckFatal(err)

		logging.CheckFatal(cmd.Start())

		go io.Copy(os.Stdout, stdout)
		go io.Copy(os.Stderr, stderr)

		logging.CheckFatal(cmd.Wait())

		if !cmd.ProcessState.Success() {
			fmt.Fprintf(os.Stderr, "exec failed. stopping.")
			os.Exit(0)
		}

		took := time.Since(start)
		if took < *limit {
			fmt.Fprintf(os.Stderr, "limit reached. (took %v)", took)
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Done. (took %v)\n", took)
	}

}
