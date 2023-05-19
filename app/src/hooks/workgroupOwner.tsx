
import { useRouter } from 'next/router';
import { useWallet } from '@solana/wallet-adapter-react';


import { PublicKey } from "@solana/web3.js";


/**
 * @deprecated This should not be used anymore - we're selecting the workgroup pubkey directly
 */
export function useWorkgroupOwner(): PublicKey | null {
    const router = useRouter();
    const wallet = useWallet();
    const owner = router.query['owner'];
    if (owner) {
        // TODO: validate that its a real Solana PublicKey
        return new PublicKey(owner);
    }
    return wallet.publicKey;
}