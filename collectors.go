package main

import (
	"log"
	"time"

	"github.com/cloudfoundry/gosigar"
	"github.com/influxdb/influxdb-go"
)

type Serieses []*influxdb.Series

func CollectMemory(schan chan<- Serieses, sleepLen time.Duration) {
	var (
		mem  sigar.Mem
		swap sigar.Swap
	)

	serieses := make(Serieses, 2)

	for {
		// ram
		checkFatal(mem.Get())
		serieses[0] = &influxdb.Series{
			Name:    "Memory",
			Columns: []string{"System", "Total", "Used", "Free", "Cached"},
			Points: [][]interface{}{
				{"planc", mem.Total, mem.ActualUsed, mem.ActualFree, mem.ActualFree - mem.Free},
			},
		}
		log.Printf("Mem:  total[%12d] used[%12d] free[%12d]\n", mem.Total, mem.Used, mem.Free)

		// swap
		checkFatal(swap.Get())
		serieses[1] = &influxdb.Series{
			Name:    "Swap",
			Columns: []string{"System", "Total", "Used", "Free"},
			Points: [][]interface{}{
				{"planc", swap.Total, swap.Used, swap.Free},
			},
		}
		log.Printf("Swap: total[%12d] used[%12d] free[%12d]\n", swap.Total, swap.Used, swap.Free)

		schan <- serieses
		time.Sleep(sleepLen)
	}

}

func CollectDiskSpace(schan chan<- Serieses, sleepLen time.Duration, path string) {
	var diskspace sigar.FileSystemUsage

	for {
		checkFatal(diskspace.Get(path))
		schan <- Serieses{
			&influxdb.Series{
				Name:    "DiskSpace",
				Columns: []string{"System", "Total", "Used", "Free"},
				Points: [][]interface{}{
					{"planc", diskspace.Total * 1024, diskspace.Used * 1024, diskspace.Free * 1024},
				},
			},
		}

		log.Printf("df[%-10s]: %4s %4s %4s %4s\n", path,
			formatSize(diskspace.Total),
			formatSize(diskspace.Used),
			formatSize(diskspace.Avail),
			sigar.FormatPercent(diskspace.UsePercent()),
		)

		time.Sleep(sleepLen)
	}
}

func CollectCPULoad(schan chan<- Serieses, sleepLen time.Duration) {
	var avg sigar.LoadAverage

	for {
		checkFatal(avg.Get())
		schan <- Serieses{
			&influxdb.Series{
				Name:    "Load",
				Columns: []string{"System", "One", "Five", "Fifteen"},
				Points: [][]interface{}{
					{"planc", avg.One, avg.Five, avg.Fifteen},
				},
			},
		}

		log.Printf("Load: 1m[%.2f] 5m[%.2f] 15m[%.2f]\n", avg.One, avg.Five, avg.Fifteen)

		time.Sleep(sleepLen)
	}

}
