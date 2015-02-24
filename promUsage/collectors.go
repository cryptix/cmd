package main

import (
	"log"
	"time"

	"github.com/cloudfoundry/gosigar"
	"github.com/prometheus/client_golang/prometheus"
)

func CollectRam(sleepLen time.Duration, name string) {
	var mem sigar.Mem
	go func() {
		for {
			checkFatal(mem.Get())
			log.Printf("Mem:  total[%12d] used[%12d] free[%12d]\n", mem.Total, mem.Used, mem.Free)
			time.Sleep(sleepLen)
		}
	}()
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem:   "ram",
			Name:        "actualUsed",
			Help:        "ram that is used",
			ConstLabels: prometheus.Labels{"host": name},
		},
		func() float64 {
			return float64(mem.ActualUsed)
		},
	))
}

func CollectSwap(sleepLen time.Duration, name string) {
	var swap sigar.Swap
	go func() {
		for {
			// swap
			checkFatal(swap.Get())
			log.Printf("Swap: total[%12d] used[%12d] free[%12d]\n", swap.Total, swap.Used, swap.Free)
			time.Sleep(sleepLen)
		}
	}()
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem:   "swap",
			Name:        "used",
			Help:        "swap space that is used",
			ConstLabels: prometheus.Labels{"host": name},
		},
		func() float64 {
			return float64(swap.Used)
		},
	))
}

func CollectUsagePercent(name, path string) {
	var diskspace sigar.FileSystemUsage
	log.Printf("collecting DiskSpace[%s]", path)
	usage := func() float64 {
		checkFatal(diskspace.Get(path))
		log.Printf("df[%-10s]: %4s %4s %4s %4s\n", path,
			formatSize(diskspace.Total),
			formatSize(diskspace.Used),
			formatSize(diskspace.Avail),
			sigar.FormatPercent(diskspace.UsePercent()),
		)
		return diskspace.UsePercent()
	}
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "diskspace",
			Name:      "percentage",
			Help:      "disk space used in percent",
			ConstLabels: prometheus.Labels{
				"host": name,
				"path": path,
			},
		},
		usage,
	))
}

func CollectUsed(sleepLen time.Duration, name, path string) {
	var diskspace sigar.FileSystemUsage
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "diskspace",
		Name:      "bytesUsed",
		Help:      "disk space used in percent",
		ConstLabels: prometheus.Labels{
			"host": name,
			"path": path,
		},
	})
	prometheus.MustRegister(g)
	for {
		checkFatal(diskspace.Get(path))
		g.Set(float64(diskspace.Used * 1024))
		time.Sleep(sleepLen)
	}
}

func CollectCPULoad(sleepLen time.Duration, name string) {
	var avg sigar.LoadAverage
	go func() {
		for {
			checkFatal(avg.Get())
			log.Printf("Load: 1m[%.2f] 5m[%.2f] 15m[%.2f]\n", avg.One, avg.Five, avg.Fifteen)
			time.Sleep(sleepLen)
		}
	}()
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "usage",
			Name:      "loadavg",
			Help:      "system load",
			ConstLabels: prometheus.Labels{
				"host": name,
				"time": "1m",
			},
		},
		func() float64 { return avg.One },
	))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "usage",
			Name:      "loadavg",
			Help:      "system load",
			ConstLabels: prometheus.Labels{
				"host": name,
				"time": "5m",
			},
		},
		func() float64 { return avg.Five },
	))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "usage",
			Name:      "loadavg",
			Help:      "system load",
			ConstLabels: prometheus.Labels{
				"host": name,
				"time": "15m",
			},
		},
		func() float64 { return avg.Fifteen },
	))
}
