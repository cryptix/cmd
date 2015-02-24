package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cloudfoundry/gosigar"
	"github.com/codegangsta/cli"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	app := cli.NewApp()
	app.Name = "promUsage"
	app.Usage = "exposes statistics to prometheus"
	app.Action = run
	dfDefault := cli.StringSlice([]string{"/"})
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "host", Value: "127.0.0.1:8090", Usage: "our host to listen on for prometheus"},
		cli.StringFlag{Name: "name, n", Value: "undefined", Usage: "the namespace to use"},
		cli.DurationFlag{Name: "interval,i", Value: time.Minute},
		cli.BoolFlag{Name: "verbose,vv", Usage: "print gathered stats to stderr"},
		cli.StringSliceFlag{Name: "df", Value: &dfDefault},
	}
	app.Run(os.Args)
}

func run(ctx *cli.Context) {
	log.SetOutput(ioutil.Discard)
	if ctx.Bool("verbose") {
		log.SetOutput(os.Stderr)
	}
	n := ctx.String("name")
	i := ctx.Duration("interval")
	go CollectRam(i, n)
	go CollectSwap(i, n)
	go CollectCPULoad(i, n)
	for _, p := range ctx.StringSlice("df") {
		go CollectUsagePercent(n, p)
		go CollectUsed(i, n, p)
	}

	http.Handle("/metrics", prometheus.Handler())
	checkFatal(http.ListenAndServe(ctx.String("host"), nil))
}

// utilities
func checkFatal(err error) {
	if err != nil {
		log.SetOutput(os.Stderr) // might be ioutil.Discard
		log.Fatal(err)
	}
}

func formatSize(size uint64) string {
	return sigar.FormatSize(size * 1024)
}
