import Split from "react-split";
import { Stack } from "react-bootstrap";

import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { useQuery } from "@tanstack/react-query";
import { getProvider, getWorknetGroup } from "../worknet/worknet";

import { DeviceList } from "../components/DeviceList";
import { SpecsList } from "../components/SpecsList";
import { PublicKeyView } from "../components/PublicKeyView";
import { DeviceAdd } from "../components/DeviceAdd";
import { NewSpec } from "../components/NewSpec";
import { PublicKey } from "@solana/web3.js";
import { SpecImportFromLibraryButton } from "./SpecImportFromLibraryPopup";
import { realmsGovernanceFromWorkgroup } from "../worknet/governanceHelpers";
import { RealmsProposer } from "../worknet/createWorkgroup";
import { useProposer } from "../hooks/proposer";
import { WorkgroupSettings } from "./WorkgroupSettings";

interface WorkGroupStatusProps {
    workgroupKey: PublicKey;
}

export function WorkGroupStatus({ workgroupKey }: WorkGroupStatusProps) {
    const wallet = useWallet();
    const { connection } = useConnection();
    const provider = getProvider(connection, wallet);
    const { proposer, setProposer } = useProposer();

    const {
        status,
        data: group,
        error,
    } = useQuery(["workgroup", workgroupKey.toString()], () =>
        getWorknetGroup(provider, workgroupKey)
    );

    const { error: govError } = useQuery(
        [
            "governance",
            group?.chain?.PubKey.toString(),
            wallet.publicKey?.toString(),
        ],
        () => realmsGovernanceFromWorkgroup(provider, group),
        {
            onSuccess: (governance) => {
                if (!governance || !wallet || !wallet.publicKey || !group)
                    return;
                if (!wallet.publicKey.equals(group.groupAuthority)) {
                    setProposer(new RealmsProposer(governance, provider));
                }
            },
        }
    );

    // TODO: Integrate govError instead of just logging it
    console.log({ govError, error });

    if (group === undefined || group.chain === undefined) {
        return <div>loading...</div>;
    }

    const err = error as Error;
    if (
        status === "error" &&
        err.message.startsWith("Account does not exist")
    ) {
        // TODO: first give the user the tools to make a new WorkGroup (cos i'm on a localnet :)
        // Later, add create multi-sig or realm to then own the work group.
        const walletKey = wallet.publicKey;
        if (walletKey) {
            return (
                <Stack className="flex-fill almost-vh-100">
                    <div>
                        Worknet for {workgroupKey.toString()} not created yet
                    </div>
                </Stack>
            );
        }
        return (
            <Stack className="flex-fill almost-vh-100">
                <div>Worknet for {workgroupKey.toString()} not created yet</div>
                <div></div>
            </Stack>
        );
    }
    if (!group) {
        var errorMsg = "";
        const err = error as Error;
        if (err) {
            errorMsg = err.message; // ok
        } else {
            errorMsg = JSON.stringify(error);
        }
        return (
            <Stack className="flex-fill almost-vh-100">
                <div>
                    query: getWorknetGroup {status}: {errorMsg}
                </div>
            </Stack>
        );
    }

    if (!provider.wallet.publicKey) {
        return <div>Please connect a wallet to continue.</div>;
    }

    console.log({ proposer });

    return (
        <>
            <Stack className="flex-fill almost-vh-100">
                <div className="d-flex justify-content-between">
                    <div>
                        <h2>
                            Workgroup: {group.name} -{" "}
                            <PublicKeyView publicKey={group.chain?.PubKey} />
                        </h2>
                    </div>
                    <WorkgroupSettings workgroupKey={workgroupKey} />
                </div>
                <div className="mb-1">
                    {proposer && (
                        <>
                            <small>
                                <strong className="ms-1">Authority</strong>{" "}
                                <code className="ms-1 me-1">
                                    {group.groupAuthority.toString()}
                                </code>
                            </small>
                            {proposer.dao() && "(DAO)"}
                        </>
                    )}
                </div>
                {/* TODO: need a UI for when there is a newly created group with no devices or specs... */}
                <Split
                    sizes={[30, 70]}
                    direction="horizontal"
                    className="split"
                    gutterSize={5}
                >
                    <div className="m-2 overflow-auto">
                        <div className="d-flex align-items-center">
                            <h5 className="m-0 p-1">Devices</h5>
                            <DeviceAdd />
                        </div>
                        
                        <DeviceList
                            provider={provider}
                            devices={group.devices}
                        />
                    </div>
                    <div className="m-2 overflow-auto">
                        <div className="d-flex align-items-center">
                            <h4 className="p-1 mb-0">Specs</h4>
                            <div className="p-1">
                                <SpecImportFromLibraryButton />
                                <NewSpec provider={provider} />
                            </div>
                        </div>
                        <SpecsList
                            provider={provider}
                            specs={group.specs}
                            deployments={group.deployments}
                            devices={group.devices}
                        />
                    </div>
                </Split>
            </Stack>
        </>
    );
}

export default WorkGroupStatus;
