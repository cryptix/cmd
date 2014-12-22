// pinghealth takes a hostfile from the command line, pings each of them until the process is killed.
//
// timout, wait between pings and retries can be configured with flags.
//
// metrics are logged to influxdb using https://github.com/rcrowley/go-metrics
//
// ping construction is done by https://github.com/erikh/ping
package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/cryptix/go/backoff"
	"github.com/cryptix/go/logging"
	"github.com/erikh/ping"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/influxdb"
)

var (
	// flags
	retry   = flag.Int("r", 10, "number of retries before aborting")
	timeout = flag.Duration("t", 100*time.Millisecond, "what timeout to use for each ping")
	wait    = flag.Duration("w", 500*time.Millisecond, "how long to wait between ping bursts")

	// globals
	log      = logging.Logger("pingmon")
	timeouts metrics.Counter
)

type pinger struct {
	ip    *net.IPAddr
	name  string
	done  chan struct{}
	timer metrics.Timer
}

func NewPinger(s string) (*pinger, error) {
	var (
		err error
		p   pinger
	)

	p.name = s

	log.Debugf("Resolving for %s", s)
	p.ip, err = net.ResolveIPAddr("ip6", s)
	if err != nil {
		log.Warningf("%15s - ResolveIPAddr(ipv6) failed - %s", p.name, err)
		p.ip, err = net.ResolveIPAddr("ip4", s)
		if err != nil {
			return nil, fmt.Errorf("%15s - ResolveIPAddr(ipv4) failed - %s", p.name, err)
		}
	}

	p.done = make(chan struct{})

	p.timer = metrics.NewTimer()
	err = metrics.Register("ping."+strings.Replace(p.name, ".", "-", -1), p.timer)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *pinger) run() {
	var (
		start   time.Time
		attempt int
	)

	start = time.Now()
	for {

		select {

		case <-time.After(backoff.Default.Duration(attempt)):
			if attempt > *retry {
				log.Warningf("%15s - attempts exceeded", p.name)
				attempt = 0
				time.Sleep(1 * time.Minute)
				continue
			}

			err := ping.Pinger(p.ip, *timeout+backoff.Default.Duration(attempt))
			if err != nil { // retry
				attempt++
				log.Noticef("%15s - %50s (attempt %d | took %v)",
					p.name,
					err,
					attempt,
					time.Since(start))
				timeouts.Inc(1)
				continue
			}
			p.timer.UpdateSince(start)
			time.Sleep(*wait)

			attempt = 0
			start = time.Now() // reset after sucessfull ping - timer updates include timeout duration

		case <-p.done: // quit
			return
		}
	}
}

func main() {
	flag.Parse()

	var (
		err   error
		hostf *os.File
		hosts []*pinger
	)

	if len(flag.Args()) != 1 {
		log.Warning("No hostsfile to ping. quiting.")
		os.Exit(1)
	}

	if flag.Args()[0] == "-" {
		hostf = os.Stdin
	} else {
		hostf, err = os.Open(flag.Args()[0])
		logging.CheckFatal(err)
		defer hostf.Close()
	}

	shutdown := make(chan os.Signal)
	done := make(chan struct{})
	signal.Notify(shutdown, os.Interrupt, os.Kill)

	go func() {
		for sig := range shutdown {
			log.Warningf("captured %v, stopping pingers and exiting..", sig)
			for _, h := range hosts {
				close(h.done)
			}
			close(done)
		}
	}()

	go influxdb.Influxdb(metrics.DefaultRegistry, *timeout/2, &influxdb.Config{
		Host:     "127.0.0.1:8086",
		Database: "nethealth",
		Username: "higgs",
		Password: "logg",
	})

	timeouts = metrics.NewCounter()
	err = metrics.Register("ping.timeouts", timeouts)
	logging.CheckFatal(err)

	hostSc := bufio.NewScanner(hostf)

	for hostSc.Scan() {
		h, err := NewPinger(hostSc.Text())
		logging.CheckFatal(err)
		hosts = append(hosts, h)
	}

	logging.CheckFatal(hostSc.Err())

	for _, h := range hosts {
		go h.run()
	}

	<-done
}
