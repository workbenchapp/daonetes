
import { Cluster, clusterApiUrl } from '@solana/web3.js';
import { useRouter } from 'next/router';


export function useValidatorNetwork(): string {
    const router = useRouter();
    const network = router.query['network'];
    if (network) {
        const singleNetwork = network.toString();
        try {
            const networkUrl = clusterApiUrl(singleNetwork as Cluster);
            if (networkUrl) return networkUrl;
        } catch {
            // not one of the well known aliases.continue
            try {
                const u = new URL(singleNetwork)
                return u.toString()
            } catch (error) {
                console.log("not a valid url" + error)
            }
        }
    }

    return clusterApiUrl("devnet");
}

export function useValidatorNetworkName(): string {
    const network = useValidatorNetwork()

    //console.log("NETWORK: " + network)

    // TODO: make this more reliable - parse the url - avoid the slashes etc

    switch (network) {
        case 'https://api.devnet.solana.com':
            return 'devnet'
        case 'https://api.mainnet-beta.solana.com/':
            return 'mainnet-beta'
        case 'https://api.testnet.solana.com':
            return 'testnet'
        default:
            break;
    }

    return network;
}