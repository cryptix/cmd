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
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cryptix/go/backoff"
	"github.com/cryptix/go/logging"
	"github.com/erikh/ping"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/errgo.v1"
)

var (
	// flags
	listenAddr = flag.String("http", "localhost:1337", "on what http address to listen")
	name       = flag.String("name", "undefined", "name to give your metrics")

	retry   = flag.Int("r", 10, "number of retries before aborting")
	timeout = flag.Duration("t", 100*time.Millisecond, "what timeout to use for each ping")
	wait    = flag.Duration("w", 500*time.Millisecond, "how long to wait between ping bursts")

	// globals
	log = logging.Logger("pingmon")
)

type pinger struct {
	ip       *net.IPAddr
	host     string
	done     chan struct{}
	timeouts prometheus.Counter
	latency  prometheus.Summary
}

func NewPinger(s string) (*pinger, error) {
	var (
		err error
		p   pinger
	)
	if len(s) < 0 {
		return nil, errgo.New("host cant be empty")
	}

	p.host = s

	p.ip, err = net.ResolveIPAddr("ip6", s)
	if err != nil {
		log.Warningf("%15s - ResolveIPAddr(ipv6) failed - %s", p.host, err)
		p.ip, err = net.ResolveIPAddr("ip4", s)
		if err != nil {
			return nil, errgo.Notef(err, "ResolveIPAddr(ipv4) failed - %s", p.host)
		}
	}

	p.timeouts = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pingmon_timeouts",
		Help: "number of timeouts occured",
		ConstLabels: prometheus.Labels{
			"name": *name,
			"host": p.host,
		},
	})
	if err = prometheus.Register(p.timeouts); err != nil {
		return nil, errgo.Notef(err, "Register of timeouts failed for %s", p.host)
	}

	p.latency = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "pingmon_latency",
		Help: "how big is the latency in avaerage",
		ConstLabels: prometheus.Labels{
			"name": *name,
			"host": p.host,
		},
	})
	if err = prometheus.Register(p.latency); err != nil {
		return nil, errgo.Notef(err, "Register of latency failed for %s", p.host)
	}

	p.done = make(chan struct{})
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
				log.WithFields(logrus.Fields{
					"attempt": attempt,
					"host":    p.host,
				}).Warning("attempts exceeded")
				attempt = 0
				time.Sleep(1 * time.Minute)
				continue
			}

			err := ping.Pinger(p.ip, *timeout+backoff.Default.Duration(attempt))
			if err != nil { // retry
				attempt++
				log.WithFields(logrus.Fields{
					"host":    p.host,
					"error":   err,
					"attempt": attempt,
					"took":    time.Since(start),
				}).Info("ping failed")
				p.timeouts.Inc()
				continue
			}
			p.latency.Observe(time.Since(start).Seconds())
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

	go func() {
		lis, err := net.Listen("tcp", *listenAddr)
		logging.CheckFatal(err)
		http.Serve(lis, prometheus.Handler())
	}()

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
