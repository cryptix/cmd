package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/cloudfoundry/gosigar"
	"github.com/codegangsta/cli"
	"github.com/influxdb/influxdb-go"
)

var name string

func main() {
	app := cli.NewApp()
	app.Name = "influxUsage"
	app.Usage = "Sends usage reports to influxdb"
	app.Action = run

	app.Flags = []cli.Flag{
		cli.StringFlag{"host", "127.0.0.1:8086", "influxdb host to connect to"},
		cli.StringFlag{"name, n", "", "the name of the system to report as"},
		cli.BoolFlag{"verbose,vv", "print gathered stats to stderr"},
	}

	app.Run(os.Args)
}

func run(ctx *cli.Context) {
	// default is discard
	log.SetOutput(ioutil.Discard)
	if ctx.Bool("verbose") {
		log.SetOutput(os.Stderr)
	}

	// set default name if flag is empty
	name = ctx.String("name")
	if name == "" {
		name = "undefined"
	}

	cfg := influxdb.ClientConfig{
		Host:     ctx.String("host"),
		Database: "usage",
	}
	schan, err := NewInfluxCollector(&cfg)
	checkFatal(err)

	go CollectMemory(schan, 1*time.Second)
	go CollectCPULoad(schan, 1*time.Second)
	go CollectDiskSpace(schan, 1*time.Minute, "/")

	// lazy block..
	done := make(chan struct{})
	<-done

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
