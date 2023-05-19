package proxy

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"time"

	"github.com/gagliardetto/solana-go"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/go-logr/logr"
	"github.com/portto/solana-go-sdk/types"
	"github.com/workbenchapp/worknet/daoctl/lib/networking/dns"
	"github.com/workbenchapp/worknet/daoctl/lib/networking/ice"
	netproxy "github.com/workbenchapp/worknet/daoctl/lib/networking/proxy"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/anchor/generated/worknet"
	"github.com/workbenchapp/worknet/daoctl/lib/workgroup"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

// TODO: this will ultimately be caddy with its l4 support

// PROXY anything local to the wg network(s)
// PROXY anything on the wg network(s) to the local ips

func init() {
	proxiedDevices = make(map[string]*ProxyDevice)
}

// TODO: getDeviceList should move to something solana
func getDeviceList(ctx context.Context) []ag_solanago.PublicKey {
	devices := make([]ag_solanago.PublicKey, 0)
	log := logr.FromContextOrDiscard(ctx)

	group := workgroup.GetCachedWorkGroupInfo()
	if group == nil {
		log.Info("No workgroupinfo cached")
		return devices
	}
	if len(group.Devices) == 0 {
		log.Info("workgroupinfo devices list empty")
	}
	// Can't filter out the "deleted" devices - the index in the list is used for the meshIP
	return group.Devices
}

// And this is then the proxy bit
type ProxyDevice struct {
	Info *worknet.Device
	//DeviceKey        ag_solanago.PublicKey
	ProxyAddress     string
	WireguardAddress string
	WireguardPeerKey string

	WireguardListeners  map[string]interface{}
	LocalProxyListeners map[string]string
}

// TODO: is the key the device key, or the deviceATA key? (this drives me batty)
type ProxyDeviceList = map[string]*ProxyDevice

var proxiedDevices ProxyDeviceList

func knownDevice(device ag_solanago.PublicKey) bool {
	_, ok := proxiedDevices[device.String()]
	return ok
}

func getDeviceProxyInfo(ctx context.Context, device ag_solanago.PublicKey, idx int) {
	log := logr.FromContextOrDiscard(ctx)
	// TODO: only recreate if needed...
	info, err := workgroup.GetDeviceInfoByKey(ctx, device)
	if err != nil {
		return
	}
	if info.Status != worknet.DeviceStatusRegistered || info.Hostname == "" {
		// device not registered yet..
		return
	}
	proxyAddress := fmt.Sprintf("127.1.0.%d", idx+2)
	// TODO: need a good place to put this magic
	// TODO: Windows is argh! https://stackoverflow.com/questions/7535060/powershell-how-to-create-network-adapter-loopback
	// https://github.com/PlagueHO/LoopbackAdapter
	// https://gbe0.com/posts/windows/server-windows/create-loopback-interface-with-powershell/
	// https://www.powershellgallery.com/packages/LoopbackAdapter/1.1.0.30/Content/LoopbackAdapter.psm1
	// https://superuser.com/questions/1270789/how-to-setup-ip-alias-to-local-network-on-windows
	// MMMM netstat -b -n -a     suggests to me i'm wrong, and maybe i just broke something.
	if runtime.GOOS == "darwin" {
		c := exec.Command("ifconfig", "lo0", "alias", proxyAddress)
		err := c.Run()
		if err != nil {
			log.Error(err, "Failed to configure proxy network alias")
		}
	}
	proxiedDevices[device.String()] = &ProxyDevice{
		Info: info,
		//DeviceKey:           device,
		ProxyAddress:        proxyAddress,
		WireguardAddress:    fmt.Sprintf("192.169.99.%d", idx+2),
		WireguardListeners:  make(map[string]interface{}),
		LocalProxyListeners: make(map[string]string),
	}
	dns.UpdateDnsHostRecord(info.Hostname, net.ParseIP(proxyAddress).To4())
	dns.UpdateDnsHostRecord(device.String(), net.ParseIP(proxyAddress).To4())
}

func setDeviceOff(deviceKey string) {
	device, ok := proxiedDevices[deviceKey]
	if ok {
		device.WireguardPeerKey = "no"
		proxiedDevices[deviceKey] = device
	}
}

/*
	separate ice proxy from wireguard
	then change wireguard loop to only add nodes that we have a connection to (tunnel only initially)
	add a ProxyDevice interface / functions like:
		IsConnected
		IsLocal / GetLocalDevice
	and then re-write the wireguard bit as a goroutine that waits on an update/retry channel , and have a 'check status loop', etc
*/

func ProxyToDevices(ctx context.Context, deviceAuthorityWallet *types.Account, deviceInfoListenAddress string) {
	log := logr.FromContextOrDiscard(ctx)
	var localDevice *ProxyDevice

	deviceKeys := getDeviceList(ctx)
	for idx, deviceKey := range deviceKeys {
		if deviceKey.String() == "11111111111111111111111111111111" {
			continue // skip deleted devices
		}
		// TODO: add a timeout in case the chain info changes,
		if !knownDevice(deviceKey) {
			log.Info("Found new device", "name", deviceKey)
			getDeviceProxyInfo(ctx, deviceKey, idx)
		}
	}

	// try making tunnels
	for _, deviceKey := range deviceKeys {
		if deviceKey.String() == "11111111111111111111111111111111" {
			continue // skip deleted devices
		}
		// TODO: this is to proxy any requests to 127.1.0.x to the wireguard ip's
		device, ok := proxiedDevices[deviceKey.String()]
		if ok {
			if device.Info.DeviceAuthority.Equals(solana.PublicKey(deviceAuthorityWallet.PublicKey)) {
				continue // don't make a connection to yourself, its naf.
			}
			deviceAddr := device.ProxyAddress + ":12913" // need to proxy on a different port
			// TODO: deviceKey.String()+"Client" should really besomething else (like the localDevice..).
			ice.MakeNewICEConnectionRequest(ctx, deviceKey.String()+"Client", device.Info.DeviceAuthority.String()+"Server", deviceAddr)
			log.V(1).Info(
				"Connecting to remote wireguard using:",
				"deviceAddr", deviceAddr,
				"deviceHostname", device.Info.Hostname,
				"deviceAuthority", device.Info.DeviceAuthority.String(),
			)
		}
	}

	// TODO: extract to evented, which needs proxiedDevices to be a safe cache
	UpdateWireGuardNetwork(ctx, deviceAuthorityWallet, proxiedDevices)

	// make remove device things available here
	for _, deviceKey := range deviceKeys {
		if deviceKey.String() == "11111111111111111111111111111111" {
			continue // skip deleted devices
		}
		// TODO: this is to proxy any requests to 127.1.0.x to the wireguard ip's
		pDev, ok := proxiedDevices[deviceKey.String()]
		if !ok || pDev.Info == nil {
			continue
		}
		if pDev.Info.Status != worknet.DeviceStatusRegistered {
			continue
		}
		//spew.Dump("device", device.String(), pDev)
		if pDev.WireguardPeerKey == "no" {
			continue
		}
		// TODO: need a way to say - skip that, its busted
		if deviceKey.String() == "CyBoP8fbmLBzJLsnUtB6SAgCB1m1AE6gXwyKYe2Emy69" {
			continue // skip xeon-daolet, its not actually on
		}
		if pDev.Info.DeviceAuthority.Equals(solana.PublicKey(deviceAuthorityWallet.PublicKey)) {
			localDevice = pDev
		}
		// TODO: this should be "foreach non-local device's active deployment"
		// TODO: 9495 is a cli option - not a constant!
		ListenAndServeFromWireguard(ctx, pDev.Info.Hostname+"-deviceAPI", wireguardNet, pDev, 9495)
		deviceInfo := workgroup.GetCachedDeviceStatusInfo(pDev.Info.DeviceAuthority.String())
		if deviceInfo == nil {
			// And try to prime the deployments cache so we can create the port mappings
			UpdateDeviceInfoFromMesh(ctx, pDev)
			deviceInfo = workgroup.GetCachedDeviceStatusInfo(pDev.Info.DeviceAuthority.String())
		}
		if deviceInfo == nil {
			log.V(2).Info(
				"Skipping, no device info cached yet",
				"hostname", pDev.Info.Hostname,
				"device ATA", pDev.Info.DeviceAuthority.String(),
			)
		} else {
			for deploymentKey, info := range deviceInfo.DeployState {
				for stateKey, state := range info.States {
					log.V(2).Info("Setting up proxy", "deployment", deploymentKey, "state", stateKey)
					for _, publish := range state.Publishers {
						log.V(2).Info("Listening on port", "name", publish.Name, "protocol", publish.Protocol, "port", publish.PublishedPort)
						if publish.PublishedPort > 0 {
							ListenAndServeFromWireguard(ctx, publish.Name, wireguardNet, pDev, publish.PublishedPort)
						}
					}
				}
			}
		}
	}

	// Expose the thigns running on the local device
	ListenToWireguardAndServeFromLocalDeployments(ctx, wireguardNet, localDevice, deviceInfoListenAddress, 9495) // so the other devices can talk to local 9495
	localDeviceInfo := workgroup.GetCachedDeviceStatusInfo("")
	if localDeviceInfo == nil {
		return
	}

	// TODO: this should be "foreach localDevice's active deployment"
	for deploymentHash, deployment := range localDeviceInfo.DeployState {
		log.V(2).Info("Looking at deployment", "deployment hash", deploymentHash)

		for _, state := range deployment.States {
			for _, publish := range state.Publishers {
				log.V(2).Info("Setup proxy for", "deployment", deployment, "name", publish.Name)
				localUrl := fmt.Sprintf("%s:%d", publish.URL, int(publish.PublishedPort))

				if publish.PublishedPort > 0 {
					ListenToWireguardAndServeFromLocalDeployments(ctx, wireguardNet, localDevice, localUrl, publish.PublishedPort)
				}
			}
		}
	}
}

func GetProxyDeviceInfoByName(name string) *ProxyDevice {
	// TODO: if name=="" we mean local..
	for _, pDev := range proxiedDevices {
		if pDev.Info != nil && pDev.Info.Status == worknet.DeviceStatusRegistered {
			// match by hostname (human useful)
			if pDev.Info.Hostname == name {
				return pDev
			}
			// match by deviceATA
			if pDev.Info.DeviceAuthority.String() == name {
				return pDev
			}
			// match by device key
			// if pDev.DeviceKey.String() == name {
			// 	return pDev
			// }
		}
	}
	return nil
}

// ListenAndServe should add a listener for each port on each device to the
// 127.1.0.x range, that then talks to the wireguard tun
func ListenAndServeFromWireguard(ctx context.Context, deploymentName string, tnet *netstack.Net, pDev *ProxyDevice, deploymentPort int) {
	log := logr.FromContextOrDiscard(ctx)
	// localPort := deploymentPort
	// TODO: should do a bump if there's a clash with an existing port __maybe__
	// if localPort <= 1024 {
	// 	// TODO: only do this if we're not admin (OR if we detect the port is in use?)
	// 	localPort = localPort + 3333
	// }
	localAddr := fmt.Sprintf("%s:%d", pDev.ProxyAddress, deploymentPort)
	if tnet == nil {
		// wg not ready yet
		log.V(1).Info("no wg net", "node", localAddr)
		return
	}

	if pDev.Info == nil {
		// TODO: why am i here
		log.V(1).Info("pDev info missing", "node", localAddr)
		return
	}

	if _, ok := pDev.LocalProxyListeners[localAddr]; ok {
		return
	}
	// This will become variable
	remoteDeployAddress := fmt.Sprintf("%s:%d", pDev.WireguardAddress, deploymentPort)

	// TODO: no, this is not where we should know the dns domain...
	localDNSAddr := fmt.Sprintf("%s.%s:%d", pDev.Info.Hostname, "dmesh", deploymentPort)

	updateEndpointProxyInfo(deploymentName, deploymentPort, localDNSAddr)
	pDev.LocalProxyListeners[localAddr] = deploymentName
	go func() {
		const beNice = 1 * time.Second

		for stop := false; !stop; {
			select {
			case <-ctx.Done():
				delete(pDev.LocalProxyListeners, localAddr)
				stop = true

			default:
				// TODO: could also consider just deleteing and allowing the next pollInterval to re-init it...
				log.Info("================ Listen and Serve from Wireguard",
					"remote", pDev.Info.Hostname,
					"remote addr", remoteDeployAddress,
					"local addr", localAddr,
					"local dns", localDNSAddr,
					"device ATA", pDev.Info.DeviceAuthority.String(),
				)

				netproxy.ForwardTCPToMesh(ctx, localAddr, localDNSAddr, remoteDeployAddress, tnet)
				time.Sleep(time.Duration(beNice))
			}
		}
	}()
}

// This is the listener for the wireguard ports that should then request to the local deployment
func ListenToWireguardAndServeFromLocalDeployments(ctx context.Context, tnet *netstack.Net /*port, handler, idk*/, pDev *ProxyDevice, localDeploymentAddress string, wgDeploymentPort int) {
	log := logr.FromContextOrDiscard(ctx)
	wireguardListenAddr := fmt.Sprintf(":%d", wgDeploymentPort)
	if tnet == nil {
		// wg not ready yet
		log.V(1).Info("no wg net", "node", wireguardListenAddr)
		return
	}

	if pDev.Info == nil {
		// TODO: why am i here
		log.V(1).Info("no pDev info", "node", wireguardListenAddr)
		return
	}

	if _, ok := pDev.WireguardListeners[wireguardListenAddr]; ok {
		return
	}

	pDev.WireguardListeners[wireguardListenAddr] = true
	go func() {
		const beNice = 1 * time.Second

		for stop := false; !stop; {
			select {
			case <-ctx.Done():
				delete(pDev.WireguardListeners, wireguardListenAddr)
				stop = true
			default:
				log.V(1).Info(
					">>>>>>>>> ListenToWireguardAndServeFromLocalDeployments",
					"wg addr", pDev.WireguardAddress,
					"wg port", wgDeploymentPort,
					"local addr", localDeploymentAddress,
				)

				// TODO: could also consider just deleteing and allowing the next pollInterval to re-init it...
				netproxy.ReceiveFromMesh(ctx, localDeploymentAddress, wgDeploymentPort, tnet)
				time.Sleep(time.Duration(beNice))
			}
		}
	}()

}
