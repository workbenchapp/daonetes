import { useState } from "react";

import { Row } from "react-bootstrap";
import { useQuery } from "@tanstack/react-query";
import { getDeployment } from "../worknet/worknet";
import { DeploymentType } from "../worknet/types";
import { SpecView } from "./SpecView";
import { AnchorProvider } from "@project-serum/anchor";
import { PublicKey } from "@solana/web3.js";
interface SpecsListProps {
    provider: AnchorProvider;
    specs: PublicKey[];
    deployments: PublicKey[];
    devices: PublicKey[];
}

export function SpecsList({
    provider,
    specs,
    deployments,
    devices,
}: SpecsListProps) {
    const [onChainDeployments, setOnChainDeployments] = useState<
        DeploymentType[]
    >([]);
    const { status, data, error } = useQuery(
        ["deployments", deployments],
        async () => {
            var specDataList: DeploymentType[] = [];
            await deployments.map(async (specKey) => {
                const specData = (await getDeployment(
                    provider,
                    specKey
                )) as DeploymentType;
                specData.chain = {
                    PubKey: specKey,
                };
                console.log("DEPLOYMENT " + specData);
                specDataList.push(specData);
            });
            setOnChainDeployments(specDataList);
            return specDataList;
        }
    );

    if (!data) {
        var errorMsg = "";
        if (error && error instanceof Error) {
            errorMsg = error.message; // ok
        } else {
            errorMsg = JSON.stringify(error);
        }
        return (
            <div>
                query: getDevice {status}: {errorMsg}
            </div>
        );
    }

    // console.log("DEPLOYMENTS: " + JSON.stringify(onChainDeployments))

    // TODO: not enough info on chain about the spec
    // TODO: how to go from spec publicKey to deployment PDA
    // TODO: how to get from spec PublicKey to devices that have a deployment?

    return (
        <Row md={7} className="g-4">
            {specs.map((specKey) => {
                return (
                    <SpecView
                        provider={provider}
                        key={specKey.toString()}
                        specKey={specKey}
                        deployments={onChainDeployments}
                        devices={devices}
                    />
                );
            })}
        </Row>
    );
}
