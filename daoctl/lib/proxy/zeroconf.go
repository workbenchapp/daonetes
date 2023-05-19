package proxy

import (
	"context"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	"github.com/grandcat/zeroconf"
)

var mdnsEntries map[string]*zeroconf.ServiceEntry

func init() {
	mdnsEntries = make(map[string]*zeroconf.ServiceEntry)
}

func QueryHostnameForIP(hostname string) *net.IP {
	fullname := fmt.Sprintf("%s._daonetes._tcp.local.", hostname)
	entry, ok := mdnsEntries[fullname]
	if !ok {
		return nil
	}
	return &entry.AddrIPv4[0] // TODO: ok, this is dumb
}

func ServeMDNS(ctx context.Context, name string) {
	// ALWAYS listen to port 9495 on the wireguard network
	server, err := zeroconf.Register(name, "_daonetes._tcp", "local.", 9495, []string{"txtv=0", "lo=1", "la=2"}, nil)
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	<-ctx.Done()

	cLog := logr.FromContextOrDiscard(ctx)
	cLog.Info("Zeroconf Shutting down.")
}

func ResolveMDNS(ctx context.Context) {
	cLog := logr.FromContextOrDiscard(ctx)

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		cLog.Error(err, "Failed to initialize resolver:")
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			// TODO: if we detect more or changes while we're running, we should re-configure the network
			cLog.Info("MDNS entry found", "mdnsHostname", entry.ServiceInstanceName(), "mdnsPort", entry.Port, "mdnsIPV4", entry.AddrIPv4[0].String())
			mdnsEntries[entry.ServiceInstanceName()] = entry
		}
		cLog.Info("MDNS: No more entries.")
	}(entries)

	err = resolver.Browse(ctx, "_daonetes._tcp", "local.", entries)
	if err != nil {
		cLog.Error(err, "Failed to browse")
	}

	<-ctx.Done()
}
