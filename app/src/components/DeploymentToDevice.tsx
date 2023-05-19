import React, { useState } from "react";
import { Button, Modal } from "react-bootstrap";
import Form from "react-bootstrap/Form";

import { scheduleDeployment } from "../worknet/createWorkgroup";
import { useQuery } from "@tanstack/react-query";
import { getDevice } from "../worknet/worknet";
import { PublicKey } from "@solana/web3.js";
import { AnchorProvider } from "@project-serum/anchor";
import { DeploymentType } from "../worknet/types";
import { useProposer } from "../hooks/proposer";
import { toastAndSendProposedTransactions } from "../worknet/displays";
import { useWorkgroup } from "../hooks/selectedWorkgroup";

interface DeviceOptionProps {
    provider: AnchorProvider;
    deviceKey: PublicKey;
}
function DeviceOption({ provider, deviceKey }: DeviceOptionProps) {
    const { status, data, error } = useQuery(
        ["device", deviceKey],
        async () => {
            const devicePk = new PublicKey(deviceKey);
            return await getDevice(provider, devicePk);
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

    const device = data;
    // const DEVICE = {
    //     "ipv4": [59, 167, 214, 59],
    //     "hostname": "p1",
    //     "bump": 253,
    //     "status": { "registered": {} },
    //     "deviceAuthority": "2YuBYJ93yJXZkoxTWXJZkFR5SfRXvPiBGWc48CFePxgw",
    //     "groupAuthority": "9tVLKVDWFD6CnZdat2QM1v9ev4673u28GDrbNVsMRGYi"
    // };
    if (!device.hostname) {
        return <></>;
    }
    return (
        <option value={device.deviceAuthority.toString()}>
            {device.hostname}
        </option>
    );
}

interface DeploymentToDeviceProps {
    provider: AnchorProvider;
    deployment: DeploymentType;
    devices: PublicKey[];
    showing: boolean;

    // TODO: WTF
    // eslint-disable-next-line no-unused-vars
    setShow: (s: boolean) => void;
}
export function DeploymentToDevice({
    provider,
    deployment,
    devices,
    showing,
    setShow,
}: DeploymentToDeviceProps) {
    const [deviceKey, setDeviceKey] = useState<undefined | string>("");
    const handleClose = () => setShow(false);
    const workgroupKey = useWorkgroup();
    const { proposer } = useProposer();

    return (
        <>
            {showing && (
                <Modal show={showing} onHide={handleClose}>
                    <Modal.Header closeButton>
                        <Modal.Title>Schedule</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <div>
                            <small className="text-secondary">
                                Deploy {deployment.name} to the selected device
                            </small>
                            <Form>
                                <Form.Group
                                    className="mb-3"
                                    controlId="addSpecForm.ControlInput1"
                                >
                                    <Form.Select
                                        aria-label="select device"
                                        onChange={(e) =>
                                            setDeviceKey(e.target.value)
                                        }
                                    >
                                        {!deviceKey && (
                                            <option value={undefined}>
                                                Select
                                            </option>
                                        )}
                                        {devices.map((devKey) => {
                                            return (
                                                <DeviceOption
                                                    provider={provider}
                                                    key={devKey.toString()}
                                                    deviceKey={devKey}
                                                />
                                            );
                                        })}
                                    </Form.Select>
                                    {deviceKey && (
                                        <div>
                                            <small className="font-weight-bold text-secondary">
                                                Device Key:
                                            </small>
                                            <code>{deviceKey}</code>
                                        </div>
                                    )}
                                </Form.Group>
                            </Form>
                        </div>
                    </Modal.Body>
                    <Modal.Footer>
                        <Button
                            disabled={!deviceKey}
                            onClick={async () => {
                                if (!deviceKey || !proposer || !workgroupKey) {
                                    handleClose();
                                    throw {
                                        msg: "Expected deviceKey, proposer, and workgroupKey to be defined",
                                        vals: {
                                            deviceKey,
                                            proposer,
                                            workgroupKey,
                                        },
                                    };
                                }
                                toastAndSendProposedTransactions(
                                    "Schedule",
                                    proposer,
                                    provider.connection,
                                    await scheduleDeployment(
                                        provider,
                                        proposer,
                                        workgroupKey,
                                        deployment.name,
                                        new PublicKey(deviceKey)
                                    )
                                );
                                handleClose();
                            }}
                        >
                            {proposer?.dao() && "Propose "}Schedule
                        </Button>
                    </Modal.Footer>
                </Modal>
            )}
        </>
    );
}
