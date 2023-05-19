export async function getLocalhostDeviceInfo(hostname?: string) {
    if (!hostname || hostname === "") {
        hostname = "";
    }
    // {
    //     "deviceInfo": {
    //       "Ipv4": [
    //         59,
    //         167,
    //         214,
    //         59
    //       ],
    //       "Hostname": "p1",
    //       "Bump": 252,
    //       "Status": 1,
    //       "DeviceAuthority": "72wkUiH3ahMbQaefnHgL16UJVrpcyWwUdLYGvZDgCKf2",
    //       "WorkGroup": "3DAWMFkYskFmTfWiD5yTu9ELzoPW1tWspQi6Mp9S3eek"
    //     },
    //     "deviceInfoKey": "1UZ6Jw93oNq1p2tnF5RyaQMzWGH8ozEEE6VdphQUnRG",
    //     "deviceInfoKeyBump": "252",
    //     "deviceWallet": "72wkUiH3ahMbQaefnHgL16UJVrpcyWwUdLYGvZDgCKf2",
    //     "validator": "https://api.devnet.solana.com"
    //   }
    var url = new URL("http://localhost:9495/device")
    url.searchParams.set("proxy", hostname)
    const response = await fetch(url.toString());
    return await response.json();
}

// returns a map of "deploymenthash:port" to "localproxy:port"
export async function getLocalhostEndpointInfo() {
    var url = new URL("http://localhost:9495/endpoints")
    const response = await fetch(url.toString());
    return response.json() as Promise<Map<string, string>>;
}