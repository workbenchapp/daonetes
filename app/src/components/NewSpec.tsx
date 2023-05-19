import { AnchorProvider } from "@project-serum/anchor";
import React, { useState } from "react";
import { Button, Modal } from "react-bootstrap";
import Form from "react-bootstrap/Form";
import { useProposer } from "../hooks/proposer";
import { useWorkgroup } from "../hooks/selectedWorkgroup";

import { createSpec } from "../worknet/createWorkgroup";
import { toastAndSendProposedTransactions } from "../worknet/displays";

interface NewSpecProps {
    provider: AnchorProvider;
}
export function NewSpec({ provider }: NewSpecProps) {
    const workGroup = useWorkgroup();
    const [show, setShow] = useState(false);
    const [specURL, setSpecURL] = useState("");
    const [specName, setSpecName] = useState(""); // TODO: generate a random suffix
    const { proposer } = useProposer();

    const handleClose = () => setShow(false);
    const handleShow = () => setShow(true);

    return (
        <>
            <Button variant="secondary" onClick={handleShow}>
                Add
            </Button>

            <Modal centered show={show} onHide={handleClose}>
                <Modal.Header closeButton>
                    <Modal.Title>Add Spec</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <div>
                        <Form>
                            <Form.Group
                                className="mb-3"
                                controlId="addSpecForm.ControlInput1"
                            >
                                <Form.Label>Spec Name</Form.Label>
                                <Form.Control
                                    type="text"
                                    placeholder="e.g., Gumdrop indexer"
                                    value={specName}
                                    onChange={(e) =>
                                        setSpecName(e.target.value)
                                    }
                                />
                            </Form.Group>
                            <Form.Group
                                className="mb-3"
                                controlId="addSpecForm.SpecPublicKey"
                            >
                                <Form.Label>Docker Compose YAML URL</Form.Label>
                                <Form.Control
                                    type="textarea"
                                    placeholder="https://example.com/docker-compose.yaml"
                                    value={specURL}
                                    onChange={(e) => setSpecURL(e.target.value)}
                                />
                            </Form.Group>
                        </Form>
                    </div>
                    <hr />
                    <div></div>
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={handleClose}>
                        Close
                    </Button>
                    <Button
                        disabled={!workGroup ? true : false}
                        onClick={async () => {
                            if (!workGroup || !proposer) return;
                            toastAndSendProposedTransactions(
                                "CreateWorkSpec",
                                proposer,
                                provider.connection,
                                await createSpec(
                                    provider,
                                    proposer,
                                    workGroup,
                                    specName,
                                    specURL,
                                    "" // Metadata URL
                                )
                            );
                            handleClose();
                        }}
                    >
                        {proposer?.dao() && "Propose "}Register Spec
                    </Button>
                </Modal.Footer>
            </Modal>
        </>
    );
}
