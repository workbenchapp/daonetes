import { AnchorProvider } from "@project-serum/anchor";
import { PublicKey } from "@solana/web3.js";
import { useQuery } from "@tanstack/react-query";
import { MintDataType } from "../worknet/types";
import { getAttachedTokenMints, getDeployment } from "../worknet/worknet";

interface ScheduledDeploymentProps {
    provider: AnchorProvider;
    mint: MintDataType;
}
export function ScheduledDeployment({
    provider,
    mint,
}: ScheduledDeploymentProps) {
    const { status, data, error } = useQuery(
        ["deployment", mint.data.parsed.info.mintAuthority],
        async () => {
            const deploymentKey = new PublicKey(
                mint.data.parsed.info.mintAuthority
            );
            return await getDeployment(provider, deploymentKey);
        }
    );

    if (!data) {
        var errorMsg = "";
        if (error && error instanceof Error) {
            errorMsg = error.message; // ok
        } else {
            errorMsg = JSON.stringify(error);
        }
        console.log(`query: getDeployment ${status}: ${errorMsg}`);
        return (
            <li>
                query: getDeployment {status}: {errorMsg}
            </li>
        );
    }
    console.log("Deployment: " + JSON.stringify(data));

    return <li>{data.name}</li>;
}

interface ScheduledDeploymentsProps {
    provider: AnchorProvider;
    dKey: PublicKey;
}
export function ScheduledDeployments({
    provider,
    dKey,
}: ScheduledDeploymentsProps) {
    const { status, data, error } = useQuery(
        ["scheduled", dKey.toString()],
        async () => {
            // TODO: also get the number of tokens from the ATA
            return await getAttachedTokenMints(provider, dKey);
        }
    );

    if (!data) {
        var errorMsg = "";
        if (error && error instanceof Error) {
            errorMsg = error.message; // ok
        } else {
            errorMsg = JSON.stringify(error);
        }
        console.log(`query: getAttachedTokenMints ${status}: ${errorMsg}`);
        return (
            <div>
                query: getAttachedTokenMints {status}: {errorMsg}
            </div>
        );
    }
    console.log("MINTLIST: " + JSON.stringify(data));

    return (
        <div>
            <div>scheduled.. {data.length} deployments</div>
            <ul>
                {data.map((mint) => {
                    return (
                        <ScheduledDeployment
                            provider={provider}
                            key={mint.toString()}
                            mint={mint}
                        />
                    );
                })}
            </ul>
        </div>
    );
}
