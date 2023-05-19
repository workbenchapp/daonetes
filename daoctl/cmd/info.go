package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/proxy"
	"github.com/workbenchapp/worknet/daoctl/lib/workgroup"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type InfoCmd struct {
	Format string `help:"Output format: [default, json, spew]" default:"default" yaml:"format"`
	Show   string `help:"Show info about: [device, network]" default:"network" yaml:"show"`
	Node   string `help:"request network infor from selected node (use 127.1.0.x)" default:"localhost" yaml:"Node"`
}

func GetProxyDeviceInfoByDeviceWallet(deviceWallet string, proxiedDevices map[string]*proxy.ProxyDevice) (string, *proxy.ProxyDevice) {
	for deviceKey, device := range proxiedDevices {
		if device.Info.DeviceAuthority.String() == deviceWallet {
			return deviceKey, device
		}
	}
	return "", nil
}

func (r *InfoCmd) Run(gOpts *options.GlobalOptions) error {
	ctx := gOpts.Ctx
	var result interface{}
	var err error
	switch r.Show {
	case "device":
		fmt.Printf("Device info:\n\n") // From Solana - this assumes we have access to the on disk device wallet, and other crimes.
		result, err = workgroup.GetDeviceInfo(ctx)
	case "network":
		// TODO: iterate through all nodes, and show who's connected to whom
		url := fmt.Sprintf("http://%s:9495/wireguard", r.Node)
		//gOpts.Log.Info("Get", "url", url)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("ERROR-Get: %s\n", err)
			r.Show = "device"
			return r.Run(gOpts)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("ERROR-ReadAll: %s\n", err)
			r.Show = "device"
			return r.Run(gOpts)
		}
		if resp.StatusCode != 200 {
			fmt.Printf("ERROR(%s): %s\n", resp.Status, body)
		}
		var device proxy.NetworkStatusAPIInfo
		err = json.Unmarshal(body, &device)
		if err != nil {
			fmt.Printf("ERROR-UnMarshal: %s\n", err)
			r.Show = "device"
			return r.Run(gOpts)
		}
		fmt.Printf("Network info:\n\n")
		// TODO: output the name of the workgroup that we're using...
		//       BUT the name onchain can be different to the one in the local config file
		result = device
		if r.Format == "default" {
			nodename := ""
			localDeviceATA := ""
			LocalAddress := ""
			_, localProxyDevice := GetProxyDeviceInfoByDeviceWallet(device.DeviceWallet, device.ProxyDevices)
			if localProxyDevice != nil {
				nodename = localProxyDevice.Info.Hostname
				localDeviceATA = localProxyDevice.Info.DeviceAuthority.String()
				LocalAddress = localProxyDevice.ProxyAddress
			}
			// TODO:  device key
			fmt.Printf("NodeName:   %s\n", nodename)
			fmt.Printf("DeviceATA:  %s\n", localDeviceATA)
			fmt.Printf("WG PeerKey: %s\n", device.PublicKey.String())
			fmt.Printf("Endpoint:   %s\n", LocalAddress)
			//fmt.Printf("    Endpoint:\t%s\n", device.IP.String())
			fmt.Printf("ListenPort: %d\n", device.ListenPort)
			if localProxyDevice != nil {
				for address, deploymentName := range localProxyDevice.LocalProxyListeners {
					// TODO: replace ip with dns name
					fmt.Printf("Service:    %s -> %s\n", deploymentName, address)
				}
			}

			// TODO: sort consistently - nodename would be nice
			keys := make([]string, 0)
			keyedOutput := make(map[string]string)
			for deviceKey, proxyDevice := range device.ProxyDevices {
				if device.DeviceWallet == proxyDevice.Info.DeviceAuthority.String() {
					continue // this is the "local device" above
				}
				nodename := proxyDevice.Info.Hostname
				deviceATA := proxyDevice.Info.DeviceAuthority.String()

				// find peer by matching
				var peer wgtypes.Peer
				//TODO: yeah, if we're going to go through it more than once, convert to map?
				for _, p := range device.Peers {
					if proxyDevice.ProxyAddress == p.Endpoint.IP.String() {
						peer = p
					}
				}
				wgPublicKey := ""
				peerAllowedIps := ""
				peerEndPointIp := ""
				peerLastSeen := "never"
				if peer.Endpoint == nil || proxyDevice.ProxyAddress != peer.Endpoint.IP.String() {
					//continue // wg not connected to it yet, so it's not in the wg cfg list - very like due to it being off...
				} else {
					wgPublicKey = peer.PublicKey.String()
					for _, ip := range peer.AllowedIPs {
						peerAllowedIps = peerAllowedIps + ip.IP.String()
					}
					peerEndPointIp = peer.Endpoint.IP.String()
					timeSince := time.Since(peer.LastHandshakeTime)
					if timeSince > 5*peer.PersistentKeepaliveInterval {
						peerLastSeen = peer.LastHandshakeTime.String()
					} else {
						peerLastSeen = timeSince.Round(time.Second).String()
					}
				}

				output := ""
				output = output + fmt.Sprintf("  NodeName:\t%s\n", nodename)
				output = output + fmt.Sprintf("    DeviceATA:\t%s\n", deviceATA)
				output = output + fmt.Sprintf("    WG PeerKey:\t%s\n", wgPublicKey)
				output = output + fmt.Sprintf("    MeshIP:\t %s\n", peerAllowedIps)
				output = output + fmt.Sprintf("    Endpoint:\t%s\n", peerEndPointIp)
				//output = output+fmt.Sprintf("  KeepAlive:  %s\n", peer.PublicKey.String())
				output = output + fmt.Sprintf("    LastSeen:\t~%s\n", peerLastSeen)
				output = output + fmt.Sprintf("    Rx:\t\t%d\t\t Tx: %d\n", peer.ReceiveBytes, peer.TransmitBytes)
				//output = output+fmt.Sprintf("  Protocol:  %s\n", peer.)

				iceStatus := "unknown"
				iceKey := fmt.Sprintf("%sClient_%sServer", deviceKey, deviceATA) // TODO: move this into the ice code.. (and it should be localDeviceATA)
				if iceState, ok := device.IceConnection[iceKey]; ok {
					iceStatus = iceState.Status
				}
				output = output + fmt.Sprintf("    ice state:\t%s\n", iceStatus)

				if proxyDevice != nil {
					for address, deploymentName := range proxyDevice.LocalProxyListeners {
						// TODO: replace ip with dns name
						output = output + fmt.Sprintf("    Service:\t%s -> %s\n", deploymentName, address)
					}
				}
				keyedOutput[nodename] = output
				keys = append(keys, nodename)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Print(keyedOutput[key])
			}
			return nil
		}
	}
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return err
	}

	switch r.Format {
	case "spew":
		spew.Dump(result)
	case "json", "default":
		b, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
		fmt.Printf("%s\n", string(b))
	}

	return nil
}
