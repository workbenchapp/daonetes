import { AnchorProvider } from "@project-serum/anchor";
import React, { useState } from "react";
import { Button, Modal } from "react-bootstrap";
import Form from "react-bootstrap/Form";
import { useProposer } from "../hooks/proposer";
import { useWorkgroup } from "../hooks/selectedWorkgroup";

import { createDeployment } from "../worknet/createWorkgroup";
import { toastAndSendProposedTransactions } from "../worknet/displays";
import { SpecType } from "../worknet/types";

interface DeploymentButtonProps {
    provider: AnchorProvider;
    spec: SpecType;
}
export function DeploymentButton({ provider, spec }: DeploymentButtonProps) {
    const workgroupKey = useWorkgroup();
    const { proposer } = useProposer();

    const [show, setShow] = useState(false);
    const [deploymentName, setDeploymentName] = useState("my-deployment"); // TODO: generate a random suffix
    const [replicas, setReplicas] = useState("1");

    const handleClose = () => setShow(false);
    const handleShow = () => setShow(true);

    return (
        <>
            <Button variant="secondary" onClick={handleShow}>
                Add
            </Button>

            <Modal centered show={show} onHide={handleClose}>
                <Modal.Header closeButton>
                    <div className="flex-column">
                        <Modal.Title>Create Deployment Mint</Modal.Title>
                        <div className="text-secondary">{spec.name}</div>
                    </div>
                </Modal.Header>
                <Modal.Body>
                    <div>
                        <Form>
                            <Form.Group
                                className="mb-3"
                                controlId="addSpecForm.ControlInput1"
                            >
                                <Form.Label className="fw-bold">
                                    Deployment Name
                                </Form.Label>
                                <Form.Control
                                    type="text"
                                    value={deploymentName}
                                    onChange={(e) =>
                                        setDeploymentName(e.target.value)
                                    }
                                />
                                <Form.Label className="fw-bold">
                                    Deployment Replicas
                                </Form.Label>
                                <Form.Control
                                    type="text"
                                    value={replicas}
                                    onChange={(e) =>
                                        setReplicas(e.target.value)
                                    }
                                />
                            </Form.Group>
                        </Form>
                    </div>
                    <hr />
                    <div></div>
                </Modal.Body>
                <Modal.Footer>
                    <Button
                        onClick={async () => {
                            if (!proposer || !workgroupKey)
                                throw {
                                    msg: "Cannot transact, proposer or workgroupKey undefined",
                                    proposer,
                                    workgroupKey: workgroupKey?.toString(),
                                };
                            if (deploymentName.length === 0) {
                                throw "Deployment must have name";
                            }
                            toastAndSendProposedTransactions(
                                "CreateDeployment",
                                proposer,
                                provider.connection,
                                await createDeployment(
                                    provider,
                                    proposer,
                                    workgroupKey,
                                    spec.name,
                                    deploymentName,
                                    Number.parseInt(replicas, 10)
                                )
                            );
                            handleClose();
                        }}
                    >
                        {proposer?.dao() && "Propose "}Create
                    </Button>
                </Modal.Footer>
            </Modal>
        </>
    );
}
