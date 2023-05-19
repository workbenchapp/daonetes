//go:build darwin

package dns

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// OSX didnt' like putting the dns service on 127.1.0.1 :/
var dnsAddress = "127.0.0.1"

// TODO: oh wow, need to run
// sudo ifconfig lo0 alias 127.1.0.x
// for each device

func EnsureDNSConfigured() error {
	domain := "dmesh"
	resolverDir := "/etc/resolver"
	os.MkdirAll(resolverDir, 0755)

	// if it already exists, check it or leave it.

	resolvedConf := filepath.Join(resolverDir, domain) //"/etc/resolver/dmesh"
	linesToWrite := fmt.Sprintf(`
domain %s
nameserver %s
search_order 1
timeout 5
	`, domain, dnsAddress)

	err := ioutil.WriteFile(resolvedConf, []byte(linesToWrite), 0644)
	if err != nil {
		return err
	}

	// TODO: nees to reset the DNS services to make sure things get noticed.
	// https://phoenixnap.com/kb/how-to-flush-dns-cache

	return nil
}
