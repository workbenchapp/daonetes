package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/workbenchapp/worknet/daoctl/lib/networking/ice"
	"github.com/workbenchapp/worknet/daoctl/lib/networking/wgctl"
	"github.com/workbenchapp/worknet/daoctl/lib/solana"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type NetworkStatusAPIInfo struct {
	wgtypes.Device
	ProxyDevices  ProxyDeviceList
	IceConnection map[string]ice.Status
	DeviceWallet  string
}

// service+port map to dns+port map for webUI (key = composedeploymentname:port) (value = dnsaddress:port)
var endpointProxyInfo map[string]string

func updateEndpointProxyInfo(deploymentName string, port int, localAddress string) {
	key := fmt.Sprintf("%s:%d", deploymentName, port)
	endpointProxyInfo[key] = localAddress
}
func initAPIHandlers() {
	endpointProxyInfo = make(map[string]string)

	AddAPIHandler("/endpoints", func(w http.ResponseWriter, r *http.Request) {
		var replyBytes []byte

		replyBytes, err := json.MarshalIndent(endpointProxyInfo, "", "  ")
		if err != nil {
			w.Header().Set("ProxyInfo-Marshal-Error", err.Error())
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(replyBytes)
	})
	AddAPIHandler("/ice", func(w http.ResponseWriter, r *http.Request) {
		// TODO: report on ice nat-traversal status

		w.Header().Set("Content-Type", "application/json")
		status := ice.GetConnectionStates()

		//spew.Fdump(w, device)
		specJSON, err := json.MarshalIndent(status, "", " ")
		if err != nil {
			spew.Fdump(w, err)
			return
		}
		w.Write(specJSON)
	})

	// Get what wireguard-go's current state is
	AddAPIHandler("/wireguard", func(w http.ResponseWriter, r *http.Request) {
		// TODO: make another endpoint for debug log
		// TODO: make another endpoint for the API we send into wg-go
		// TODO: annotate each node with the wg Peer String() so the wg-go debug logs make sense

		w.Header().Set("Content-Type", "application/json")
		//IpComment := fmt.Sprintf("# MY IP: %s\n", wireguardAddress)
		//w.Write([]byte(IpComment))
		var b bytes.Buffer
		if wireguardDev == nil {
			w.Header().Set("Error", "Not ready yet")
			return
		}
		wireguardDev.IpcGetOperation(&b)
		device, err := wgctl.ParseDevice(&b)
		if err != nil {
			spew.Fdump(w, err)
			return
		}

		// TODO: yeah, puke, let's not add a dep to solana here
		ourWallet, err := solana.MustGetAgentWallet(r.Context())
		if err != nil {
			spew.Fdump(w, err)
			return
		}

		// Lets not transmit the private key
		device.PrivateKey = wgtypes.Key{}
		addInfo := NetworkStatusAPIInfo{
			Device:        *device,
			ProxyDevices:  proxiedDevices,
			IceConnection: ice.GetConnectionStates(),
			DeviceWallet:  ourWallet.PublicKey.String(),
		}
		//spew.Fdump(w, device)
		// TODO: be nice to elide the Keys entirely / replace with the dns name
		specJSON, err := json.MarshalIndent(addInfo, "", " ")
		if err != nil {
			spew.Fdump(w, err)
			return
		}
		w.Write(specJSON)
	})

}
