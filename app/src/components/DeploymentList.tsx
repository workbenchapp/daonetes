import { AnchorProvider, web3 } from "@project-serum/anchor";
import { ParsedAccountData, PublicKey } from "@solana/web3.js";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { Button, Dropdown, Modal } from "react-bootstrap";
import { BsThreeDots } from "react-icons/bs";
import { useProposer } from "../hooks/proposer";
import { useWorkgroup } from "../hooks/selectedWorkgroup";
import { closeDeployment } from "../worknet/createWorkgroup";
import { toastAndSendProposedTransactions } from "../worknet/displays";
import { DeploymentType, SpecType } from "../worknet/types";
import { getWorknetGroup, getWorknetProgram } from "../worknet/worknet";
import { DeploymentToDevice } from "./DeploymentToDevice";

interface DeploymentProps {
    spec: SpecType;
    deployment: DeploymentType;
    devices: PublicKey[];
    provider: AnchorProvider;
}
export function Deployment({ deployment, devices, provider }: DeploymentProps) {
    const [modalShowing, setModalShowing] = useState(false);
    const [replicaTokenAmount, setReplicaTokenAmount] = useState<
        number | undefined
    >();
    const [closingDeployment, setClosingDeployment] = useState(false);
    const { proposer } = useProposer();
    const workgroupKey = useWorkgroup();
    if (!workgroupKey) {
        throw "no workgroup selected";
    }
    const { data: group } = useQuery(
        ["workgroup", workgroupKey.toString()],
        () => getWorknetGroup(provider, workgroupKey)
    );
    useEffect(() => {
        const getParsedDeploymentTokenAccount = async () => {
            if (!deployment.chain) return;
            // eslint-disable-next-line no-unused-vars
            const [deploymentTokenAccount, _bump] =
                await web3.PublicKey.findProgramAddress(
                    [
                        deployment.chain.PubKey.toBuffer(),
                        Buffer.from("deployment_tokens"),
                    ],
                    getWorknetProgram(provider).programId
                );
            const replicaTokenAccountInfo =
                await provider.connection.getParsedAccountInfo(
                    deploymentTokenAccount
                );
            const accountData = replicaTokenAccountInfo.value
                ?.data as ParsedAccountData;
            setReplicaTokenAmount(accountData.parsed.info.tokenAmount.uiAmount);
        };
        getParsedDeploymentTokenAccount();
    }, [deployment.chain, provider, provider.connection]);
    return (
        <div className="d-flex align-items-center justify-content-between">
            <div className="flex-column">
                {deployment.name}
                <div className="text-secondary">
                    {replicaTokenAmount}/{deployment.replicas} replicas
                </div>
            </div>
            {replicaTokenAmount && replicaTokenAmount > 0 ? (
                <>
                    <div className="justify-content-end">
                        <Button
                            variant="secondary"
                            onClick={() => setModalShowing(true)}
                            className="m-2 d-inline"
                        >
                            Schedule
                        </Button>

                        <Dropdown className="d-inline">
                            <Dropdown.Toggle
                                variant="secondary"
                                id="dropdown-basic"
                            >
                                <BsThreeDots />
                            </Dropdown.Toggle>
                            <Dropdown.Menu>
                                <Dropdown.Item
                                    onClick={() => setClosingDeployment(true)}
                                    href="#/close-deployment"
                                >
                                    Close Deployment
                                </Dropdown.Item>
                            </Dropdown.Menu>
                        </Dropdown>
                        <Modal
                            onHide={() => setClosingDeployment(false)}
                            show={closingDeployment}
                        >
                            <Modal.Header closeButton>
                                <Modal.Title>Close Workgroup?</Modal.Title>
                            </Modal.Header>
                            <Modal.Body>
                                <p>
                                    This will close the account associated with
                                    this deployment.
                                </p>
                                <p>
                                    <span className="text-danger">
                                        {" "}
                                        This is a destructive action and cannot
                                        be undone.
                                    </span>
                                </p>
                                <p>Are you sure?</p>
                            </Modal.Body>
                            <Modal.Footer>
                                <Button
                                    onClick={async () => {
                                        if (
                                            !proposer ||
                                            !group ||
                                            !deployment.chain
                                        )
                                            throw "Proposer or work group or deployment key missing";

                                        await toastAndSendProposedTransactions(
                                            `CloseDeployment`,
                                            proposer,
                                            provider.connection,
                                            await closeDeployment(
                                                provider,
                                                proposer,
                                                workgroupKey,
                                                deployment.chain?.PubKey
                                            )
                                        );
                                    }}
                                    variant="danger"
                                >
                                    Close Deployment
                                </Button>
                            </Modal.Footer>
                        </Modal>
                    </div>
                    <DeploymentToDevice
                        setShow={setModalShowing}
                        showing={modalShowing}
                        provider={provider}
                        deployment={deployment}
                        devices={devices}
                    />
                </>
            ) : (
                <span></span>
            )}
        </div>
    );
}
interface DeploymentListProps {
    spec: SpecType;
    deployments: DeploymentType[];
    devices: PublicKey[];
    provider: AnchorProvider;
}
export function DeploymentList({
    spec,
    deployments,
    devices,
    provider,
}: DeploymentListProps) {
    console.log({ deployments });
    return (
        <ul className="list-group">
            <li className="list-group-item">
                {deployments.length > 0 ? (
                    deployments.map((dep) => {
                        return (
                            <Deployment
                                provider={provider}
                                devices={devices}
                                spec={spec}
                                deployment={dep}
                                key={dep.name}
                            />
                        );
                    })
                ) : (
                    <small className="text-secondary">No deployments.</small>
                )}
            </li>
        </ul>
    );
}
