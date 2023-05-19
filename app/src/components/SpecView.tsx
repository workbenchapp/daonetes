import { AnchorProvider } from "@project-serum/anchor";
import { PublicKey } from "@solana/web3.js";
import { useQuery } from "@tanstack/react-query";
import moment from "moment";
import { useState } from "react";
import { Button, Dropdown, Modal } from "react-bootstrap";
import { BsThreeDots } from "react-icons/bs";
import { useProposer } from "../hooks/proposer";
import { useWorkgroup } from "../hooks/selectedWorkgroup";
import { closeSpec } from "../worknet/createWorkgroup";
import { toastAndSendProposedTransactions } from "../worknet/displays";
import { DeploymentType } from "../worknet/types";
import { getSpec } from "../worknet/worknet";
import { DeploymentButton } from "./DeploymentButton";
import { DeploymentList } from "./DeploymentList";
interface SpecViewProps {
    provider: AnchorProvider;
    specKey: PublicKey;
    deployments: DeploymentType[];
    devices: PublicKey[];
}
export function SpecView({
    provider,
    specKey,
    deployments,
    devices,
}: SpecViewProps) {
    const workgroupKey = useWorkgroup();
    const { status, data, error } = useQuery(["spec", specKey], async () => {
        const specPk = new PublicKey(specKey);
        return await getSpec(provider, specPk);
    });
    const [closingSpec, setClosingSpec] = useState(false);
    const { proposer } = useProposer();

    if (!data) {
        var errorMsg = "";
        if (error && error instanceof Error) {
            errorMsg = error.message; // ok
        } else {
            errorMsg = JSON.stringify(error);
        }
        return (
            <div>
                query: getSpec {status}: {errorMsg}
            </div>
        );
    }

    // Filter all the workgroup's deployments to the ones that use this spec
    const deploymentsForThisSpec = deployments.filter((deployment) => {
        return deployment.spec.toString() === specKey.toString();
    });

    const spec = data;

    function formatUrlOrContents(urlOrContents: string) {
        if (urlOrContents.startsWith("https://")) {
            return (
                <a rel="noreferrer" href={urlOrContents} target="_blank">
                    Link
                </a>
            );
        }
        return (
            <pre>
                <code>urlOrContents</code>
            </pre>
        );
    }

    const createdAt = spec.createdAt.toNumber() * 1000;
    const modifiedAt = spec.modifiedAt.toNumber() * 1000;

    return (
        <div className="p-4">
            <div>
                <h5 className="d-inline">{spec.name}</h5>
                <Dropdown className="d-inline">
                    <Dropdown.Toggle variant="secondary" id="dropdown-basic">
                        <BsThreeDots />
                    </Dropdown.Toggle>
                    <Dropdown.Menu>
                        <Dropdown.Item
                            onClick={() => setClosingSpec(true)}
                            href="#/close-spec"
                        >
                            Close Spec
                        </Dropdown.Item>
                    </Dropdown.Menu>
                </Dropdown>
                <Modal onHide={() => setClosingSpec(false)} show={closingSpec}>
                    <Modal.Header closeButton>
                        <Modal.Title>Close Spec?</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <p>
                            This will close the account associated with this
                            spec.
                        </p>
                        <p>
                            <span className="text-danger">
                                {" "}
                                This is a destructive action and cannot be
                                undone.
                            </span>
                        </p>
                        <p>Are you sure?</p>
                    </Modal.Body>
                    <Modal.Footer>
                        <Button
                            onClick={async () => {
                                if (!proposer || !workgroupKey)
                                    throw "Proposer or workgroup missing";

                                await toastAndSendProposedTransactions(
                                    `CloseSpec`,
                                    proposer,
                                    provider.connection,
                                    await closeSpec(
                                        provider,
                                        proposer,
                                        workgroupKey,
                                        specKey
                                    )
                                );
                                setClosingSpec(false);
                            }}
                            variant="danger"
                        >
                            Close Spec
                        </Button>
                    </Modal.Footer>
                </Modal>
            </div>
            <dl className="row">
                <dt className="col-sm-3">Type</dt>
                <dd className="col-sm-9">{Object.keys(spec.workType)}</dd>
                <dt className="col-sm-3">Created At</dt>
                <dd className="col-sm-9 text-secondary">
                    {moment(createdAt).fromNow().toString()}
                </dd>
                <dt className="col-sm-3">Modified At</dt>
                <dd className="col-sm-9 text-secondary">
                    {moment(modifiedAt).fromNow().toString()}
                </dd>
                <dt className="col-sm-3">Url or Contents</dt>
                <dd className="col-sm-9">
                    {formatUrlOrContents(spec.urlOrContents)}
                </dd>
                <dt className="col-sm-3">SHA256</dt>
                <dd className="col-sm-9">
                    <div className="overflow-hidden">
                        <code>{spec.contentsSha256}</code>
                    </div>
                </dd>
                <dt className="col-sm-3">Metadata URL</dt>
                <dd className="col-sm-9">{spec.metadataUrl || <i>Empty</i>}</dd>
                <dt className="col-sm-3">Mutable</dt>
                <dd className="col-sm-9">{spec.mutable ? "Yes" : "No"}</dd>
            </dl>
            <div className="d-flex align-items-center">
                <h6 className="m-0">Deployment Mints</h6>
                <div className="m-2">
                    <DeploymentButton
                        provider={provider}
                        spec={spec}
                        key={specKey.toString()}
                    />
                </div>
            </div>
            <DeploymentList
                provider={provider}
                spec={spec}
                deployments={deploymentsForThisSpec}
                devices={devices}
            />
        </div>
    );
}
