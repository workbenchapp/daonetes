package dns

import (
	"context"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	"github.com/miekg/dns"
)

// global constants that likely should be cmdline options and dao-cfg (and will be interesting when supporting multiple workgroups)
var dnsPort = ":53"
var tld = "dmesh."

// global struct to store dns entries
var hostmap = map[string]net.IP{
	"local." + tld: net.ParseIP("127.0.0.1"),
}

// updateDnsInfo
func UpdateDnsHostRecord(hostname string, ip net.IP) error {
	// TODO: yeah, better to be careful about testing if its already fully qualified, if it ends in a dot, or has the tld etc
	fullname := fmt.Sprintf("%s.%s", hostname, tld)
	//fmt.Printf("--- DNS entry %s to %s\n", fullname, ip.String())
	hostmap[fullname] = ip
	return nil
}

// Installer stub that calls the OS specific implementation to tell the OS to use it...
// https://minikube.sigs.k8s.io/docs/handbook/addons/ingress-dns/

// server that listens for requests and answers them
func RunDnsService(ctx context.Context) {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("Listening to DNS queries", "addr", dnsAddress, "port", dnsPort)
	server := &dns.Server{
		Addr: dnsAddress + dnsPort,
		Net:  "udp",
	}
	dns.HandleFunc(tld, handleRequest)
	go func() {
		<-ctx.Done()
		log.Info("Conext canceled, shutting doen the DNS service")
		server.Shutdown()
	}()
	if err := server.ListenAndServe(); err != nil {
		err := fmt.Errorf("(Needs to be run as root) Failed to set udp listener %s", err.Error())
		panic(err)
	}
}

func handleRequest(w dns.ResponseWriter, request *dns.Msg) {
	reply := new(dns.Msg)
	reply.SetReply(request)
	if ip, ok := hostmap[request.Question[0].Name]; ok {
		if request.Question[0].Qtype == dns.TypeA {
			// who me, care about IPv6?
			reply.Authoritative = true
			aRec := &dns.A{
				Hdr: dns.RR_Header{
					Name:   request.Question[0].Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: ip, //net.ParseIP(fmt.Sprintf("127.1.0.%d", 11)).To4(),
			}
			reply.Answer = append(reply.Answer, aRec)
		}
	}
	w.WriteMsg(reply)
}
