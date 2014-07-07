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

func main() {
	app := cli.NewApp()
	app.Name = "influxUsage"
	app.Usage = "Sends usage reports to influxdb"
	app.Flags = []cli.Flag{
		cli.BoolFlag{"verbose,vv", "print gathered stats to stderr"},
	}
	app.Action = run

	app.Run(os.Args)
}

func run(ctx *cli.Context) {
	// default is discard
	log.SetOutput(ioutil.Discard)
	if ctx.Bool("verbose") {
		log.SetOutput(os.Stderr)
	}

	cfg := influxdb.ClientConfig{
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
