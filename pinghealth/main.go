// pinghealth takes hosts from the command line, pings each of them a number of times.
//
// timout, number of pings and retries can be configured with flags.
//
// ping construction is done by https://github.com/erikh/ping
package main

import (
	"flag"
	"net"
	"os"
	"sync"
	"time"

	"github.com/cryptix/go/backoff"
	"github.com/cryptix/go/logging"
	"github.com/erikh/ping"
	"github.com/sethgrid/multibar"
)

var (
	// flags
	cnt     = flag.Int("c", 10, "how much pings to send to each host")
	retry   = flag.Int("r", 10, "number of retries before aborting")
	timeout = flag.Duration("t", 100*time.Millisecond, "what timeout to use for each ping")

	// globals
	bars   *multibar.BarContainer
	barMap map[string]func(int)
	log    = logging.Logger("pinghealth")
)

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Warning("No hosts to ping. quiting.")
		os.Exit(1)
	}

	// construct multibars
	var err error
	bars, err = multibar.New()
	logging.CheckFatal(err)

	// need to bars.MakeBar() before bars.Listen()...
	barMap = make(map[string]func(int))
	for _, h := range flag.Args() {
		barMap[h] = bars.MakeBar(*cnt, h)
	}
	bars.Println()
	go bars.Listen()

	// start the pingers
	var wg sync.WaitGroup
	wg.Add(len(flag.Args()))
	for _, h := range flag.Args() {
		go tryPings(&wg, h)
	}

	wg.Wait()
}

func tryPings(wg *sync.WaitGroup, host string) {
	var (
		err     error
		attempt int
		ip      *net.IPAddr
	)
	defer wg.Done()

	ip, err = net.ResolveIPAddr("ip6", host)
	if err != nil {
		ip, err = net.ResolveIPAddr("ip4", host)
		if err != nil {
			bars.Printf("%15s - ResolveIPAddr() failed", host)
			return
		}
	}

	for i := 0; i < *cnt; i++ {
		barMap[host](i)
		time.Sleep(500 * time.Millisecond)

		if attempt > *retry {
			bars.Printf("%15s %2d - attempts exceeded", ip, i)
			return
		}

		start := time.Now()
		err := ping.Pinger(ip, *timeout+backoff.Default.Duration(attempt))
		if err != nil { // retry
			attempt++
			i--
			bars.Printf("%15s %2d - %50s (attempt %d | took %v)",
				host,
				i,
				err,
				attempt,
				time.Since(start))
			continue
		}

		attempt = 0
	}
	barMap[host](*cnt)
}
