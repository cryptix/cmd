package main

import (
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
	app.Action = run

	app.Run(os.Args)
}

func run(ctx *cli.Context) {
	var err error

	// sigar variables
	var (
		concreteSigar sigar.ConcreteSigar
		mem           sigar.Mem
		swap          sigar.Swap
		diskspace     sigar.FileSystemUsage
	)

	// influx varaibles
	var (
		db    *influxdb.Client
		dbCfg influxdb.ClientConfig

		series []*influxdb.Series
	)

	dbCfg.Database = "usage"

	db, err = influxdb.NewClient(&dbCfg)
	if err != nil {
		log.Fatal(err)
	}

	for {
		series = []*influxdb.Series{}
		// load avg
		avg, err := concreteSigar.GetLoadAverage()
		if err != nil {
			log.Fatal(err)
		}

		series = append(series, &influxdb.Series{
			Name:    "Load",
			Columns: []string{"System", "One", "Five", "Fifteen"},
			Points: [][]interface{}{
				{"planc", avg.One, avg.Five, avg.Fifteen},
			},
		})
		log.Printf("Load: 1m[%.2f] 5m[%.2f] 15m[%.2f]\n", avg.One, avg.Five, avg.Fifteen)

		// mem
		mem.Get()
		series = append(series, &influxdb.Series{
			Name:    "Memory",
			Columns: []string{"System", "Total", "Used", "Free"},
			Points: [][]interface{}{
				{"planc", mem.Total, mem.Used, mem.Free},
			},
		})
		log.Printf("Mem:  total[%12d] used[%12d] free[%12d]\n", mem.Total, mem.Used, mem.Free)

		// swap
		swap.Get()
		series = append(series, &influxdb.Series{
			Name:    "Swap",
			Columns: []string{"System", "Total", "Used", "Free"},
			Points: [][]interface{}{
				{"planc", swap.Total, swap.Used, swap.Free},
			},
		})
		log.Printf("Swap: total[%12d] used[%12d] free[%12d]\n", swap.Total, swap.Used, swap.Free)

		// disk space
		diskspace.Get("/")
		series = append(series, &influxdb.Series{
			Name:    "DiskSpace",
			Columns: []string{"System", "Total", "Used", "Free"},
			Points: [][]interface{}{
				{"planc", diskspace.Total, diskspace.Used, diskspace.Free},
			},
		})
		log.Printf("df/: %4s %4s %4s %4s\n",
			formatSize(diskspace.Total),
			formatSize(diskspace.Used),
			formatSize(diskspace.Avail),
			sigar.FormatPercent(diskspace.UsePercent()),
		)

		err = db.WriteSeries(series)
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(1 * time.Second)
	}
}

func formatSize(size uint64) string {
	return sigar.FormatSize(size * 1024)
}
