// pollEndpoint is a helper utility that waits for a http endpoint to be reachable and return with http.StatusOK
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cryptix/go/logging"
)

var (
	endpoint = flag.String("ep", "http://127.0.0.1:5001/version", "which http endpoint path to hit")
	tries    = flag.Int("tries", 10, "how many tries to make before failing")
	timeout  = flag.Duration("tout", time.Second, "how long to wait between attempts")

	log   logging.Interface
	check = logging.CheckFatal
)

func main() {
	flag.Parse()
	logging.SetupLogging(nil)
	log = logging.Logger("pollEndpoint")

	// construct url to dial
	u, err := url.Parse(*endpoint)
	check(err)

	// show what we got
	start := time.Now()
	log.Log("event", "starting", "ties", *tries, "timeout", *timeout, "url", u.String())

	for *tries > 0 {

		err := checkOK(http.Get(u.String()))
		if err == nil {
			log.Log("event", "reachable", "left", *tries, "took", time.Since(start))
			os.Exit(0)
		}
		log.Log("event", "reqFailed", "err", err)
		time.Sleep(*timeout)
		*tries--
	}

	log.Log("event", "failed")
	os.Exit(1)
}

func checkOK(resp *http.Response, err error) error {
	if err == nil { // request worked
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pollEndpoint: ioutil.ReadAll() Error: %s", err)
		}
		return fmt.Errorf("Response not OK. %d %s %q", resp.StatusCode, resp.Status, string(body))
	}
	return err
}
