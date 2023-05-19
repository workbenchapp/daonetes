package pubip

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jpillora/backoff"
)

// modified version of https://github.com/chyeh/pubip
// because that package had issues dealing with one provider
// returning a different result

var (
	apiURLs = []string{
		"https://api.ipify.org",
		"http://myexternalip.com/raw",
		"http://ipinfo.io/ip",
		"http://ipecho.net/plain",
		"http://icanhazip.com",
		"http://ifconfig.me/ip",
		"http://ident.me",
		"http://checkip.amazonaws.com",
		"http://bot.whatismyipaddress.com",
		"http://whatismyip.akamai.com",
		"http://wgetip.com",
		"http://ip.appspot.com",
		"http://ip.tyk.nu",
		"https://shtuff.it/myip/short",
	}
)

const (
	maxTries = 3
	timeout  = 2 * time.Second

	// How many services need to agree on the
	// public IP
	consensusThreshold = 0.5
)

func getIPFromService(dest string) (net.IP, error) {
	b := &backoff.Backoff{
		Jitter: true,
	}
	client := &http.Client{}

	req, err := http.NewRequest("GET", dest, nil)
	if err != nil {
		return nil, err
	}

	for tries := 0; tries < maxTries; tries++ {
		resp, err := client.Do(req)
		if err != nil {
			d := b.Duration()
			time.Sleep(d)
			continue
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 200 {
			return nil, errors.New(dest + " status code " + strconv.Itoa(resp.StatusCode) + ", body: " + string(body))
		}

		tb := strings.TrimSpace(string(body))
		ip := net.ParseIP(tb)
		if ip == nil {
			return nil, errors.New("IP address not valid: " + tb)
		}
		return ip, nil
	}

	return nil, errors.New("failed to reach " + dest)
}

func detailErr(err error, errs []error) error {
	errStrs := []string{err.Error()}
	for _, e := range errs {
		errStrs = append(errStrs, e.Error())
	}
	j := strings.Join(errStrs, "\n")
	return errors.New(j)
}

func validate(results []net.IP) (net.IP, error) {
	resultCounts := map[string]int{}
	if results == nil {
		return nil, fmt.Errorf("failed to get any result from %d APIs", len(apiURLs))
	}
	if len(results) < 3 {
		return nil, fmt.Errorf("less than %d results from %d APIs", 3, len(apiURLs))
	}
	for _, res := range results {
		resultCounts[res.String()] += 1
	}
	for ip, counts := range resultCounts {
		if (float64(counts) / float64(len(results))) > consensusThreshold {
			return net.ParseIP(ip), nil
		}
	}
	return nil, fmt.Errorf("no consensus on public IP, results: %#v", resultCounts)
}

func worker(d string, r chan<- net.IP, e chan<- error) {
	ip, err := getIPFromService(d)
	if err != nil {
		e <- err
		return
	}
	r <- ip
}

func IP() (net.IP, error) {
	var results []net.IP
	resultCh := make(chan net.IP, len(apiURLs))
	var errs []error
	errCh := make(chan error, len(apiURLs))

	for _, d := range apiURLs {
		go worker(d, resultCh, errCh)
	}
	for {
		select {
		case err := <-errCh:
			errs = append(errs, err)
		case r := <-resultCh:
			results = append(results, r)
		case <-time.After(timeout):
			r, err := validate(results)
			if err != nil {
				return nil, detailErr(err, errs)
			}
			return r, nil
		}
	}
}
