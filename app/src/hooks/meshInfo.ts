import { PublicKey } from "@solana/web3.js";


type wgKey = number[];
type wgPeer = {
    PublicKey: wgKey,
    PresharedKey: wgKey,
    Endpoint: {
        IP: string,
        Port: number,
        Zone: string,
    },
    PersistentKeepaliveInterval: number,
    LastHandshakeTime: string,
    ReceiveBytes: number,
    TransmitBytes: number,
    AllowedIPs: {
        IP: string,
        Mask: string,
    }[],
    ProtocolVersion: number,
}

export type deviceInfo = {
    Info: {
        Ipv4: number[],
        Hostname: string,
        Bump: number,
        Status: number,
        DeviceAuthority: PublicKey,
        WorkGroup: PublicKey
    },
    ProxyAddress: string,  //"127.1.0.4",
    WireguardAddress: string, //"192.169.99.4",
    WireguardPeerKey: string,   // TODO: need to make a type that can be inited by the string represenation here, and the number[] represenation in wgPeer OR fix wgPeer.PublicKey on the way out or the API...
    WireguardListeners: {},
    LocalProxyListeners: Map<string, string>, // AddrPort: ServiceName
}
type iceInfo = {
    Status: string, // TODO: its a string enum..
}
// TODO: this structure grew from the deploy code, which morphed into proxying, which then morphed into wg and then later ice
//       it will be replaced by something less solana->wg->ice specific
export type meshInfo = {
    // TODO: separate these into wg, device and ice? and then union??
    Name: string,   // from wgStatus
    Type: number,   // from wgStatus
    PrivateKey: wgKey,  // from wgStatus // blanked out, so can be elided...
    PublicKey: wgKey, // from wgStatus
    ListenPort: number, // from wgStatus
    FirewallMark: number,  // from wgStatus
    Peers: wgPeer[], // from wgStatus
    ProxyDevices: Map<PublicKey, deviceInfo>, // this is solana devicePublicKey, deviceinfo
    IceConnection: Map<string, iceInfo>,
}

export async function getMeshInfo() {
    var url = new URL("http://localhost:9495/wireguard")
    const response = await fetch(url.toString());
    return response.json() as Promise<meshInfo>;
}

// example.
// var data = {
//     "Name": "",
//     "Type": 0,
//     "PrivateKey": [
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0,
//         0
//     ],
//     "PublicKey": [
//         125,
//         227,
//         66,
//         17,
//         227,
//         66,
//         183,
//         116,
//         100,
//         89,
//         0,
//         128,
//         4,
//         112,
//         76,
//         135,
//         101,
//         88,
//         172,
//         251,
//         72,
//         238,
//         133,
//         50,
//         191,
//         159,
//         39,
//         234,
//         167,
//         237,
//         76,
//         127
//     ],
//     "ListenPort": 12912,
//     "FirewallMark": 0,
//     "Peers": [
//         {
//             "PublicKey": [
//                 188,
//                 70,
//                 174,
//                 103,
//                 186,
//                 174,
//                 49,
//                 253,
//                 36,
//                 67,
//                 219,
//                 220,
//                 61,
//                 239,
//                 242,
//                 17,
//                 105,
//                 68,
//                 119,
//                 23,
//                 85,
//                 236,
//                 251,
//                 124,
//                 173,
//                 170,
//                 80,
//                 247,
//                 33,
//                 0,
//                 233,
//                 105
//             ],
//             "PresharedKey": [
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0
//             ],
//             "Endpoint": {
//                 "IP": "127.0.0.1",
//                 "Port": 53688,
//                 "Zone": ""
//             },
//             "PersistentKeepaliveInterval": 25000000000,
//             "LastHandshakeTime": "2022-10-20T12:26:22.665340553+10:00",
//             "ReceiveBytes": 6044,
//             "TransmitBytes": 7400,
//             "AllowedIPs": [
//                 {
//                     "IP": "192.169.99.10",
//                     "Mask": "/////w=="
//                 }
//             ],
//             "ProtocolVersion": 1
//         },
//         {
//             "PublicKey": [
//                 57,
//                 52,
//                 214,
//                 86,
//                 171,
//                 70,
//                 152,
//                 108,
//                 195,
//                 62,
//                 243,
//                 99,
//                 29,
//                 47,
//                 195,
//                 248,
//                 170,
//                 29,
//                 40,
//                 3,
//                 3,
//                 173,
//                 29,
//                 196,
//                 74,
//                 133,
//                 180,
//                 217,
//                 104,
//                 216,
//                 93,
//                 23
//             ],
//             "PresharedKey": [
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0
//             ],
//             "Endpoint": {
//                 "IP": "127.1.0.11",
//                 "Port": 12913,
//                 "Zone": ""
//             },
//             "PersistentKeepaliveInterval": 25000000000,
//             "LastHandshakeTime": "2022-10-20T12:27:12.245321143+10:00",
//             "ReceiveBytes": 6328,
//             "TransmitBytes": 6260,
//             "AllowedIPs": [
//                 {
//                     "IP": "192.169.99.11",
//                     "Mask": "/////w=="
//                 }
//             ],
//             "ProtocolVersion": 1
//         },
//         {
//             "PublicKey": [
//                 132,
//                 135,
//                 138,
//                 23,
//                 138,
//                 152,
//                 217,
//                 232,
//                 103,
//                 136,
//                 194,
//                 246,
//                 66,
//                 201,
//                 84,
//                 179,
//                 101,
//                 176,
//                 127,
//                 102,
//                 63,
//                 58,
//                 169,
//                 194,
//                 167,
//                 155,
//                 198,
//                 13,
//                 99,
//                 228,
//                 181,
//                 2
//             ],
//             "PresharedKey": [
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0
//             ],
//             "Endpoint": {
//                 "IP": "127.1.0.12",
//                 "Port": 12913,
//                 "Zone": ""
//             },
//             "PersistentKeepaliveInterval": 25000000000,
//             "LastHandshakeTime": "2022-10-20T12:27:18.632004472+10:00",
//             "ReceiveBytes": 6040,
//             "TransmitBytes": 6100,
//             "AllowedIPs": [
//                 {
//                     "IP": "192.169.99.12",
//                     "Mask": "/////w=="
//                 }
//             ],
//             "ProtocolVersion": 1
//         },
//         {
//             "PublicKey": [
//                 41,
//                 186,
//                 226,
//                 172,
//                 248,
//                 103,
//                 230,
//                 163,
//                 73,
//                 238,
//                 89,
//                 81,
//                 56,
//                 100,
//                 65,
//                 126,
//                 235,
//                 116,
//                 218,
//                 31,
//                 149,
//                 163,
//                 166,
//                 99,
//                 183,
//                 101,
//                 218,
//                 67,
//                 207,
//                 1,
//                 124,
//                 106
//             ],
//             "PresharedKey": [
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0
//             ],
//             "Endpoint": {
//                 "IP": "127.1.0.13",
//                 "Port": 12913,
//                 "Zone": ""
//             },
//             "PersistentKeepaliveInterval": 25000000000,
//             "LastHandshakeTime": "0001-01-01T00:00:00Z",
//             "ReceiveBytes": 0,
//             "TransmitBytes": 53132,
//             "AllowedIPs": [
//                 {
//                     "IP": "192.169.99.13",
//                     "Mask": "/////w=="
//                 }
//             ],
//             "ProtocolVersion": 1
//         },
//         {
//             "PublicKey": [
//                 246,
//                 178,
//                 137,
//                 4,
//                 53,
//                 168,
//                 239,
//                 118,
//                 240,
//                 217,
//                 93,
//                 48,
//                 28,
//                 194,
//                 234,
//                 74,
//                 140,
//                 82,
//                 54,
//                 191,
//                 217,
//                 147,
//                 99,
//                 97,
//                 117,
//                 252,
//                 199,
//                 177,
//                 202,
//                 89,
//                 239,
//                 90
//             ],
//             "PresharedKey": [
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0
//             ],
//             "Endpoint": {
//                 "IP": "127.0.0.1",
//                 "Port": 43831,
//                 "Zone": ""
//             },
//             "PersistentKeepaliveInterval": 25000000000,
//             "LastHandshakeTime": "2022-10-20T12:27:28.398981982+10:00",
//             "ReceiveBytes": 5592,
//             "TransmitBytes": 6260,
//             "AllowedIPs": [
//                 {
//                     "IP": "192.169.99.16",
//                     "Mask": "/////w=="
//                 }
//             ],
//             "ProtocolVersion": 1
//         },
//         {
//             "PublicKey": [
//                 217,
//                 249,
//                 39,
//                 155,
//                 112,
//                 170,
//                 23,
//                 239,
//                 61,
//                 146,
//                 21,
//                 99,
//                 121,
//                 6,
//                 235,
//                 255,
//                 184,
//                 132,
//                 217,
//                 97,
//                 124,
//                 167,
//                 240,
//                 7,
//                 181,
//                 138,
//                 106,
//                 243,
//                 241,
//                 158,
//                 108,
//                 83
//             ],
//             "PresharedKey": [
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0
//             ],
//             "Endpoint": {
//                 "IP": "127.0.0.1",
//                 "Port": 47160,
//                 "Zone": ""
//             },
//             "PersistentKeepaliveInterval": 25000000000,
//             "LastHandshakeTime": "2022-10-20T12:26:41.987531438+10:00",
//             "ReceiveBytes": 9544,
//             "TransmitBytes": 10664,
//             "AllowedIPs": [
//                 {
//                     "IP": "192.169.99.4",
//                     "Mask": "/////w=="
//                 }
//             ],
//             "ProtocolVersion": 1
//         },
//         {
//             "PublicKey": [
//                 70,
//                 185,
//                 218,
//                 99,
//                 192,
//                 44,
//                 108,
//                 72,
//                 186,
//                 254,
//                 152,
//                 18,
//                 243,
//                 169,
//                 126,
//                 224,
//                 200,
//                 192,
//                 72,
//                 227,
//                 29,
//                 243,
//                 153,
//                 21,
//                 149,
//                 47,
//                 197,
//                 96,
//                 38,
//                 233,
//                 222,
//                 81
//             ],
//             "PresharedKey": [
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0,
//                 0
//             ],
//             "Endpoint": {
//                 "IP": "127.0.0.1",
//                 "Port": 59324,
//                 "Zone": ""
//             },
//             "PersistentKeepaliveInterval": 25000000000,
//             "LastHandshakeTime": "2022-10-20T12:27:35.892137501+10:00",
//             "ReceiveBytes": 5832,
//             "TransmitBytes": 6132,
//             "AllowedIPs": [
//                 {
//                     "IP": "192.169.99.8",
//                     "Mask": "/////w=="
//                 }
//             ],
//             "ProtocolVersion": 1
//         }
//     ],
//     "ProxyDevices": {
//         "6Awab3JuaXzbhCxPRRUZYH4ox6YNCbLiKoDVmk7JrV1u": {
//             "Info": {
//                 "Ipv4": [
//                     168,
//                     138,
//                     189,
//                     219
//                 ],
//                 "Hostname": "daoctl",
//                 "Bump": 255,
//                 "Status": 1,
//                 "DeviceAuthority": "FfoAbKBTXM587pm1xp9tEiu16ga3YHqvHi4UREDcWTRE",
//                 "WorkGroup": "37z7HWZqJBppRWtym7goMhEKB6TRDembLuwQ35QALddb"
//             },
//             "ProxyAddress": "127.1.0.12",
//             "WireguardAddress": "192.169.99.12",
//             "WireguardPeerKey": "hIeKF4qY2ehniML2QslUs2Wwf2Y/OqnCp5vGDWPktQI=",
//             "WireguardListeners": {},
//             "LocalProxyListeners": {
//                 "127.1.0.12:9495": "daoctl-deviceAPI"
//             }
//         },
//         "8pa3A2UAmiZCLc6i4Uz5mhHwyxX1QyRqXWLBw5Lz1crg": {
//             "Info": {
//                 "Ipv4": [
//                     59,
//                     167,
//                     214,
//                     59
//                 ],
//                 "Hostname": "xeon",
//                 "Bump": 255,
//                 "Status": 1,
//                 "DeviceAuthority": "3bmasR4iYEaRUYJgveW8i1cddorn1Lw8cQCMVTWhiXHy",
//                 "WorkGroup": "37z7HWZqJBppRWtym7goMhEKB6TRDembLuwQ35QALddb"
//             },
//             "ProxyAddress": "127.1.0.4",
//             "WireguardAddress": "192.169.99.4",
//             "WireguardPeerKey": "2fknm3CqF+89khVjeQbr/7iE2WF8p/AHtYpq8/GebFM=",
//             "WireguardListeners": {},
//             "LocalProxyListeners": {
//                 "127.1.0.4:49153": "daonetesdzeyhxdzgmtvakvy-echo-1",
//                 "127.1.0.4:9495": "xeon-deviceAPI"
//             }
//         },
//         "AtFqc8RQ3JZSg61eyfRA1ptWcdvwmawbeZ5S2H3Yfce6": {
//             "Info": {
//                 "Ipv4": [
//                     50,
//                     206,
//                     88,
//                     162
//                 ],
//                 "Hostname": "Nathans-MacBook-Pro.local",
//                 "Bump": 255,
//                 "Status": 1,
//                 "DeviceAuthority": "2CPvR5gJLm4dG8zR4vk4sRz1XzDJ9cJYhCsk5jjodakC",
//                 "WorkGroup": "37z7HWZqJBppRWtym7goMhEKB6TRDembLuwQ35QALddb"
//             },
//             "ProxyAddress": "127.1.0.13",
//             "WireguardAddress": "192.169.99.13",
//             "WireguardPeerKey": "KbrirPhn5qNJ7llROGRBfut02h+Vo6Zjt2XaQ88BfGo=",
//             "WireguardListeners": {},
//             "LocalProxyListeners": {
//                 "127.1.0.13:9495": "Nathans-MacBook-Pro.local-deviceAPI"
//             }
//         },
//         "B1LVkEXX2KKsrxAnpSycrFszb2kpMJJzC1MYtexTHGGW": {
//             "Info": {
//                 "Ipv4": [
//                     59,
//                     167,
//                     214,
//                     59
//                 ],
//                 "Hostname": "server1.home.org.au",
//                 "Bump": 254,
//                 "Status": 1,
//                 "DeviceAuthority": "9p3y4Kb6YajLSoEYftzWVWKexHeAfEQKZpE2yx6aEqA1",
//                 "WorkGroup": "37z7HWZqJBppRWtym7goMhEKB6TRDembLuwQ35QALddb"
//             },
//             "ProxyAddress": "127.1.0.10",
//             "WireguardAddress": "192.169.99.10",
//             "WireguardPeerKey": "vEauZ7quMf0kQ9vcPe/yEWlEdxdV7Pt8rapQ9yEA6Wk=",
//             "WireguardListeners": {},
//             "LocalProxyListeners": {
//                 "127.1.0.10:9495": "server1.home.org.au-deviceAPI"
//             }
//         },
//         "BgEouvJrW2sxd3VBcce4kwede4jKekoYkWZMAM4pUjU8": {
//             "Info": {
//                 "Ipv4": [
//                     178,
//                     128,
//                     31,
//                     12
//                 ],
//                 "Hostname": "do",
//                 "Bump": 254,
//                 "Status": 1,
//                 "DeviceAuthority": "516e7RtK9CrVf3qgeMsupwoNmHnRozCj17pQepuwFRT6",
//                 "WorkGroup": "37z7HWZqJBppRWtym7goMhEKB6TRDembLuwQ35QALddb"
//             },
//             "ProxyAddress": "127.1.0.8",
//             "WireguardAddress": "192.169.99.8",
//             "WireguardPeerKey": "RrnaY8AsbEi6/pgS86l+4MjASOMd85kVlS/FYCbp3lE=",
//             "WireguardListeners": {},
//             "LocalProxyListeners": {
//                 "127.1.0.8:9495": "do-deviceAPI"
//             }
//         },
//         "BwLcwKq7HwwL2hSeL2SdDqHGsq5MMQaP3PZ2b4ZuPZ6v": {
//             "Info": {
//                 "Ipv4": [
//                     59,
//                     167,
//                     214,
//                     59
//                 ],
//                 "Hostname": "softiron",
//                 "Bump": 252,
//                 "Status": 1,
//                 "DeviceAuthority": "BFkJweHeZbLhMDuNHdhh8PKH21LEfBbxBzbitCEcboqz",
//                 "WorkGroup": "37z7HWZqJBppRWtym7goMhEKB6TRDembLuwQ35QALddb"
//             },
//             "ProxyAddress": "127.1.0.11",
//             "WireguardAddress": "192.169.99.11",
//             "WireguardPeerKey": "OTTWVqtGmGzDPvNjHS/D+KodKAMDrR3ESoW02WjYXRc=",
//             "WireguardListeners": {},
//             "LocalProxyListeners": {
//                 "127.1.0.11:9495": "softiron-deviceAPI"
//             }
//         },
//         "D8CKtmzjA289kgWJ1VpU4uggwdnWpPsrfpdxMkDTyveE": {
//             "Info": {
//                 "Ipv4": [
//                     59,
//                     167,
//                     214,
//                     59
//                 ],
//                 "Hostname": "p1",
//                 "Bump": 254,
//                 "Status": 1,
//                 "DeviceAuthority": "AijGRQyFbDsAD9yAP7GAKkLhrNfjvytyBf2JKMWLyejq",
//                 "WorkGroup": "37z7HWZqJBppRWtym7goMhEKB6TRDembLuwQ35QALddb"
//             },
//             "ProxyAddress": "127.1.0.3",
//             "WireguardAddress": "192.169.99.3",
//             "WireguardPeerKey": "",
//             "WireguardListeners": {
//                 ":49153": true,
//                 ":9495": true
//             },
//             "LocalProxyListeners": {
//                 "127.1.0.3:49153": "daonetesdzeyhxdzgmtvakvy-echo-1",
//                 "127.1.0.3:9495": "p1-deviceAPI"
//             }
//         },
//         "G595C3CNpTMTKys8v2qVraXi54DbfqmQewwKUAFCtjxZ": {
//             "Info": {
//                 "Ipv4": [
//                     59,
//                     167,
//                     214,
//                     59
//                 ],
//                 "Hostname": "x1carbon",
//                 "Bump": 254,
//                 "Status": 1,
//                 "DeviceAuthority": "5RKEf2Q39reAK5eiMCgzinaAC5D3EndnnH278ZAk3saC",
//                 "WorkGroup": "37z7HWZqJBppRWtym7goMhEKB6TRDembLuwQ35QALddb"
//             },
//             "ProxyAddress": "127.1.0.16",
//             "WireguardAddress": "192.169.99.16",
//             "WireguardPeerKey": "9rKJBDWo73bw2V0wHMLqSoxSNr/Zk2NhdfzHscpZ71o=",
//             "WireguardListeners": {},
//             "LocalProxyListeners": {
//                 "127.1.0.16:9495": "x1carbon-deviceAPI"
//             }
//         }
//     },
//     "IceConnection": {
//         "6Awab3JuaXzbhCxPRRUZYH4ox6YNCbLiKoDVmk7JrV1uClient_FfoAbKBTXM587pm1xp9tEiu16ga3YHqvHi4UREDcWTREServer": {
//             "Status": "Connected"
//         },
//         "8pa3A2UAmiZCLc6i4Uz5mhHwyxX1QyRqXWLBw5Lz1crgClient_3bmasR4iYEaRUYJgveW8i1cddorn1Lw8cQCMVTWhiXHyServer": {
//             "Status": "Connected"
//         },
//         "AtFqc8RQ3JZSg61eyfRA1ptWcdvwmawbeZ5S2H3Yfce6Client_2CPvR5gJLm4dG8zR4vk4sRz1XzDJ9cJYhCsk5jjodakCServer": {
//             "Status": "Invalid"
//         },
//         "B1LVkEXX2KKsrxAnpSycrFszb2kpMJJzC1MYtexTHGGWClient_9p3y4Kb6YajLSoEYftzWVWKexHeAfEQKZpE2yx6aEqA1Server": {
//             "Status": "Connected"
//         },
//         "BgEouvJrW2sxd3VBcce4kwede4jKekoYkWZMAM4pUjU8Client_516e7RtK9CrVf3qgeMsupwoNmHnRozCj17pQepuwFRT6Server": {
//             "Status": "Connected"
//         },
//         "BwLcwKq7HwwL2hSeL2SdDqHGsq5MMQaP3PZ2b4ZuPZ6vClient_BFkJweHeZbLhMDuNHdhh8PKH21LEfBbxBzbitCEcboqzServer": {
//             "Status": "Connected"
//         },
//         "G595C3CNpTMTKys8v2qVraXi54DbfqmQewwKUAFCtjxZClient_5RKEf2Q39reAK5eiMCgzinaAC5D3EndnnH278ZAk3saCServer": {
//             "Status": "Connected"
//         }
//     }
// }