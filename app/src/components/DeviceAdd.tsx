import React, { useState } from "react";
import { Button, Modal } from "react-bootstrap";
import Form from "react-bootstrap/Form";
import { registerDevice } from "../worknet/createWorkgroup";
import { useProposer } from "../hooks/proposer";
import { useWorkgroup } from "../hooks/selectedWorkgroup";
import { getProvider, getWorknetGroup } from "../worknet/worknet";
import { useQuery } from "@tanstack/react-query";
import { toastAndSendProposedTransactions } from "../worknet/displays";
import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { PublicKey } from "@solana/web3.js";

function isValidKey(deviceKey: string | undefined) {
    if (!deviceKey) {
        return false;
    }
    try {
        // eslint-disable-next-line no-unused-vars
        const _devicePublicKey = new PublicKey(deviceKey);
    } catch {
        return false
    }

    return true;
}

export function DeviceAdd() {
    const [show, setShow] = useState(false);
    const [deviceKey, setDeviceKey] = useState("");
    const { proposer } = useProposer();
    const workgroupKey = useWorkgroup();

    const wallet = useWallet();
    const { connection } = useConnection();
    const provider = getProvider(connection, wallet);

    // if (!workgroupKey)
    //     throw "Workgroup key should not be null when adding device";

    const { status, data: group } = useQuery(
        ["workgroup", workgroupKey?.toString()],
        () => {
            if (!workgroupKey) {
                 throw "TODO: workgroupKey not defined"
            }
            return getWorknetGroup(provider, workgroupKey);
        }
    );

    const handleClose = () => setShow(false);
    const handleShow = () => setShow(true);

    if (status != "success") {
        return (<></>);
    }

    return (
        <>
            <Button
                disabled={!isValidKey(workgroupKey?.toString())}
                className="m-2"
                size="sm"
                variant="primary"
                onClick={handleShow}
            >
                Add
            </Button>

            <Modal centered show={show} onHide={handleClose}>
                <Modal.Header closeButton>
                    <Modal.Title>Add Device to workgroup</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <div>
                        <Form>
                            <Form.Group
                                className="mb-3"
                                controlId="addDeviceForm.devicePublicKey"
                            >
                                <Form.Label>
                                    <strong>Device Key</strong>
                                </Form.Label>
                                <Form.Control
                                    type="text"
                                    placeholder="device public key"
                                    value={deviceKey}
                                    onChange={(e) =>
                                        setDeviceKey(e.target.value)
                                    }
                                />
                            </Form.Group>
                        </Form>
                    </div>
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={handleClose}>
                        Close
                    </Button>
                    <Button
                        // disable unless deviceKey is a valid formed key
                        disabled={!isValidKey(deviceKey)}
                        onClick={async () => {
                            if (!proposer || !group)
                                throw "Proposer or work group missing";

                            toastAndSendProposedTransactions(
                                "RegisterDevice",
                                proposer,
                                provider.connection,
                                await registerDevice(
                                    provider,
                                    proposer,
                                    group.groupAuthority,
                                    group.identifier,
                                    deviceKey
                                )
                            );

                            handleClose();
                        }}
                    >
                        {proposer?.dao() ? "Propose " : ""}Register Device
                    </Button>
                </Modal.Footer>
            </Modal>
        </>
    );
}
