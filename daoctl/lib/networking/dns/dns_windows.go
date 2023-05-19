//go:build windows

package dns

import (
	"fmt"
	"log"
	"os/exec"
)

var dnsAddress = "127.0.0.1"

// https://stackoverflow.com/a/54176425/31088
// dig like PowerShell: Resolve-DnsName -Server 127.0.0.1 -Name xeon.dmesh
// full help: Get-help -Name Resolve-DnsName -Full

// Tailscale chooses not to configure the DNS this way - so we continue to differentiate, not copy - https://github.com/tailscale/tailscale/blob/a12aad6b472b917daddbe1afe59e0e2745266753/net/dns/manager_windows.go#L195

// https://minikube.sigs.k8s.io/docs/handbook/addons/ingress-dns/
// PowerShell> Add-DnsClientNrptRule -Namespace ".test" -NameServers "$(minikube ip)"
// PowerShell> Get-DnsClientNrptRule | Where-Object {$_.Namespace -eq '.test'} | Remove-DnsClientNrptRule -Force; Add-DnsClientNrptRule -Namespace ".test" -NameServers "$(minikube ip)"
func EnsureDNSConfigured() error {
	removeDNS := fmt.Sprintf(`Get-DnsClientNrptRule | Where-Object {$_.Namespace -eq '.%s'} | Remove-DnsClientNrptRule -Force`, "dmesh")
	addDNS := fmt.Sprintf(`Add-DnsClientNrptRule -Namespace ".%s" -NameServers "%s"`, "dmesh", dnsAddress)

	fmt.Printf("==========================================================================\n")
	cmd := exec.Command("powershell", "-nologo", "-noprofile")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer stdin.Close()

		// TODO: should also run removeDNS on exit (on the other OS's too)
		fmt.Println("RUN: \n", removeDNS)
		fmt.Fprintln(stdin, removeDNS)

		fmt.Println("RUN: \n", addDNS)
		fmt.Fprintln(stdin, addDNS)
	}()
	out, err := cmd.CombinedOutput()
	fmt.Printf("OUTPUT: %s\n", out)

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		log.Fatal(err)
	}
	fmt.Printf("==========================================================================\n")

	return nil
}
