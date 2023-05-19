import { useRouter } from "next/router";
import { PublicKey } from "@solana/web3.js";
import { useValidatorNetworkName } from "./validatorNetwork";

export function useWorkgroup(): PublicKey | undefined {
    const router = useRouter();
    const workgroup = router.query["workgroup"];
    if (!workgroup) return undefined;
    return new PublicKey(workgroup);
}

export function useWorkgroupUrl(workgroup: PublicKey) {
    const networkName = useValidatorNetworkName();

    return getWorkgroupUrl(networkName, workgroup.toString());
}

export function getWorkgroupUrl(networkName :string, workgroup: string) {
    const newUrl = new URLSearchParams();

    newUrl.set("network", networkName);
    newUrl.set("workgroup", workgroup);

    return newUrl.toString();
}