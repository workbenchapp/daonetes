//go:build linux

package dns

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var dnsAddress = "127.1.0.1"

// TODO: really need to remove the systemd-resolv file entries when we stop the service

func EnsureDNSConfigured() error {
	//https://github.com/hashicorp/consul/pull/6731/files
	//https://systemd.network/resolved.conf.html
	resolvedConf := "/etc/systemd/resolved.conf"
	file, err := os.Open(resolvedConf)
	if err != nil {
		return fmt.Errorf("failed to open: %s", err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var text []string

	dnsline := "DNS=127.1.0.1"
	domainline := "Domains=~dmesh"

	resolvedConfChanged := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "DNS=") {
			if strings.Compare(line, dnsline) != 0 {
				resolvedConfChanged = true
				fmt.Printf("Replacing %s line (%s) with (%s)\n", resolvedConf, line, dnsline)
			}
			text = append(text, dnsline)
			// use empty string as indicator that we've written it already
			dnsline = ""
		} else if strings.HasPrefix(line, "Domains=") {
			if strings.Compare(line, domainline) != 0 {
				resolvedConfChanged = true
				fmt.Printf("Replacing /etc/systemd/resolve.conf line (%s) with (%s)\n", line, domainline)
			}
			text = append(text, domainline)
			// use empty string as indicator that we've written it already
			domainline = ""
		} else {
			text = append(text, line)
		}
	}

	if dnsline != "" {
		resolvedConfChanged = true
		fmt.Printf("Adding to /etc/systemd/resolve.conf : (%s)\n", dnsline)
		text = append(text, dnsline)
	}
	if domainline != "" {
		resolvedConfChanged = true
		fmt.Printf("Adding to /etc/systemd/resolve.conf : (%s)\n", domainline)
		text = append(text, domainline)
	}
	text = append(text, "")

	file.Close()

	if resolvedConfChanged {
		fmt.Printf("updating: %s to use %s\n", resolvedConf, domainline)

		linesToWrite := strings.Join(text, "\n")
		err = ioutil.WriteFile(resolvedConf, []byte(linesToWrite), 0644)
		if err != nil {
			return err
		}
	}

	// This is how we tell Systemd that there's a new DNS service
	// systemctl restart systemd-resolved
	stateCmd := exec.Command("systemctl", "restart", "systemd-resolved")
	stateBytes, err := stateCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to restart systemd-resolved: %s\n    (%s)", string(stateBytes), err)
	}
	// resolvectl flush-caches
	stateCmd = exec.Command("resolvectl", "flush-caches")
	stateBytes, err = stateCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to resolvectl flush-caches: %s\n    (%s)", string(stateBytes), err)
	}

	return nil
}
