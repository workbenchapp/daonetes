import { useQuery } from "@tanstack/react-query";
import { getDevice } from "../worknet/worknet";

import Container from "react-bootstrap/Container";
import Row from "react-bootstrap/Row";
import Col from "react-bootstrap/Col";

import { Accordion, Button } from "react-bootstrap";
import Badge from "react-bootstrap/Badge";

import { PublicKeyView } from "./PublicKeyView";

import { ScheduledDeployments } from "./ScheduledDeploymentsList";
import { AnchorProvider } from "@project-serum/anchor";
import { PublicKey, SystemProgram } from "@solana/web3.js";
import { DeviceType } from "../worknet/types";
import { getLocalhostDeviceInfo } from "../hooks/localhostDevice";
import { useProposer } from "../hooks/proposer";
import { useWorkgroup } from "../hooks/selectedWorkgroup";
import { toastAndSendProposedTransactions } from "../worknet/displays";
import { closeDevice } from "../worknet/createWorkgroup";

// type DeviceType = {
//     name: string;
//     publicKey: string;
//     solBalance: number;
// };

// const devices: DeviceType[] = [
//     { name: "Sven's X1Carbon", publicKey: "H1RZ5CQzusVr2cFbwHxdLEUs4fzAZ1tWjRHWoPMiJEMF", solBalance: 2 * LAMPORTS_PER_SOL },
//     { name: "Sven's Xeon", publicKey: "H1RZ5CQzusVr2cFbwHxdLEUs4fzAZ1tWjRHWoPMiJEMF", solBalance: 2 * LAMPORTS_PER_SOL },
//     { name: "Sven's P1", publicKey: "H1RZ5CQzusVr2cFbwHxdLEUs4fzAZ1tWjRHWoPMiJEMF", solBalance: 2 * LAMPORTS_PER_SOL },
//     { name: "Sven's NUC", publicKey: "H1RZ5CQzusVr2cFbwHxdLEUs4fzAZ1tWjRHWoPMiJEMF", solBalance: 2 * LAMPORTS_PER_SOL },
//     { name: "Sven's M1 mac", publicKey: "H1RZ5CQzusVr2cFbwHxdLEUs4fzAZ1tWjRHWoPMiJEMF", solBalance: 2 * LAMPORTS_PER_SOL },
// ];

// TODO: sven - learn the beauty of headless components

// enum StatusType {
//     registrationRequested,
//     registered,
//     delinquent,
//     cordoned,
// }
// type DeviceType = {
//     hostname: string;
//     ipv4: number[4];
//     bump: number;
//     status: StatusType;
//     deviceAuthority: string;
//     groupAuthority: string;
// };

type bgMap = {
    [key: string]: string;
};
interface DeviceStatusBadgeProps {
    status: string;
}
export function DeviceStatusBadge({ status }: DeviceStatusBadgeProps) {
    const bg: bgMap = {
        registered: "secondary",
        registrationRequested: "info",
        delinquent: "warning",
        cordoned: "warning",
    };
    const statusText =
        status === "registrationRequested" ? "requested" : status;
    return (
        <Badge className="fw-normal" bg={bg[status]}>
            {statusText}
        </Badge>
    );
}

interface AgentInfoProps {
    deviceInfo: any;
    deviceKey: PublicKey;
}
function AgentInfo({ deviceInfo, deviceKey }: AgentInfoProps) {
    const { status: deviceStatus, data: agentData /*, error */ } = useQuery(
        ["agentinfo", deviceInfo.hostname],
        async () => {
            try {
                const response = await getLocalhostDeviceInfo(
                    deviceInfo.hostname
                );
                console.log(
                    "getLocalhostDeviceInfo 22:" + deviceInfo.hostname,
                    { response }
                );
                return response;
            } catch (error) {
                console.log(error);
                return { error };
            }
        }
    );
    if (deviceStatus === "loading") {
        return <div>loading...</div>;
    }

    if (
        agentData &&
        agentData.deviceInfo &&
        agentData.deviceInfoKey === deviceKey.toString()
    ) {
        return (
            <Badge className="fw-normal" bg="success">
                ONLINE
            </Badge>
        );
    }

    return <></>;
}

interface WorkgroupDeviceProps {
    provider: AnchorProvider;
    deviceKey: PublicKey;
    localDeviceData: any;
}
export function WorkgroupDevice({
    provider,
    deviceKey,
    localDeviceData,
}: WorkgroupDeviceProps) {
    const { proposer } = useProposer();
    const workgroupKey = useWorkgroup();

    const { status, data, error } = useQuery(
        ["device", deviceKey],
        async () => {
            return await getDevice(provider, deviceKey);
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

    const device: DeviceType = data;
    // const DEVICE = {
    //     "ipv4": [59, 167, 214, 59],
    //     "hostname": "p1",
    //     "bump": 253,
    //     "status": { "registered": {} },
    //     "deviceAuthority": "2YuBYJ93yJXZkoxTWXJZkFR5SfRXvPiBGWc48CFePxgw",
    //     "groupAuthority": "9tVLKVDWFD6CnZdat2QM1v9ev4673u28GDrbNVsMRGYi"
    // };

    const title = device.hostname ? (
        device.hostname
    ) : (
        <PublicKeyView publicKey={device.deviceAuthority}></PublicKeyView>
    );
    const devStatus = Object.keys(device.status).join(", ");

    function replacer(key: string, value: any) {
        // Filtering out properties
        if (
            key === "ipv4" ||
            key === "chain" ||
            key === "bump" ||
            key === "hostname"
        ) {
            return undefined;
        }
        return value;
    }

    // to get what is deployed, i need to: (getAttachedTokenMints())
    // get the ATA's attached to device.deviceAuthority
    // get their mint's
    // their mint authority is the deployment data account (which also has an ATA to that same mint)
    const myLocalDeviceGroup =
        localDeviceData &&
            localDeviceData.deviceInfo &&
            localDeviceData.deviceInfoKey === device.chain?.PubKey.toString() ? (
            <Badge className="fw-normal" bg="primary">
                localDevice
            </Badge>
        ) : (
            <></>
        );

    return (
        <Accordion.Item eventKey={device.chain?.PubKey.toString() || "device"}>
            <Accordion.Header>
                <Container fluid>
                    <Row>
                        <Col md={true}>{title}</Col>
                        <Col md={4}>
                            {myLocalDeviceGroup}
                            <DeviceStatusBadge status={devStatus} />
                            {device.hostname ? (
                                <AgentInfo
                                    deviceKey={deviceKey}
                                    deviceInfo={device}
                                />
                            ) : (
                                <></>
                            )}
                        </Col>
                    </Row>
                </Container>
            </Accordion.Header>
            <Accordion.Body>
                <div>Status: {devStatus}</div>
                <div>
                    device Data: {device.chain?.PubKey.toString() || "unknown"}
                </div>
                <div>device key: {device.deviceAuthority.toString()}</div>
                <div>{device.ipv4.toString()}</div>
                <div>
                    <pre>{JSON.stringify(device, replacer, 2)}</pre>
                </div>
                {/* <div>has {dev.solBalance / LAMPORTS_PER_SOL} SOL</div> */}
                <ScheduledDeployments
                    provider={provider}
                    dKey={device.deviceAuthority}
                />
                <Button
                    variant="secondary"
                    onClick={async () => {
                        if (!proposer || !workgroupKey)
                            throw "Proposer or workgroup missing";

                        await toastAndSendProposedTransactions(
                            `CloseDevice`,
                            proposer,
                            provider.connection,
                            await closeDevice(
                                provider,
                                proposer,
                                workgroupKey,
                                deviceKey
                            )
                        );
                    }}
                >
                    Close Device
                </Button>
            </Accordion.Body>
        </Accordion.Item>
    );
}

interface DeviceListProps {
    provider: AnchorProvider;
    devices: PublicKey[];
}
export function DeviceList({ provider, devices }: DeviceListProps) {
    const { status: deviceStatus, data: localDeviceData /*, error */ } =
        useQuery(["agentinfo", "localhost"], async () => {
            return await getLocalhostDeviceInfo();
        });
    if (deviceStatus === "loading") {
        return <div>loading...</div>;
    }
    var localDevice: PublicKey | undefined;
    if (
        localDeviceData &&
        localDeviceData.deviceInfo &&
        localDeviceData.deviceInfoKey
    ) {
        localDevice = devices.find((value) => {
            if (localDeviceData.deviceInfoKey === value.toString()) {
                return true;
            }
            return false;
        });
    }

    //console.log({ devices });
    //console.log("localDevice?: " + localDevice?.toString());

    // hack for now to avoid 'invalid account discriminator' for
    // closed device accounts
    devices = devices.filter(
        (pubkey) => !pubkey.equals(SystemProgram.programId)
    );

    // TODO: list "requested" devices after "registered ones."
    return (
        <Accordion defaultActiveKey="0">
            {localDevice ? (
                <WorkgroupDevice
                    key={localDevice.toString()}
                    provider={provider}
                    deviceKey={localDevice}
                    localDeviceData={localDeviceData}
                />
            ) : (
                <></>
            )}
            {devices.map((devKey) => {
                return localDevice?.equals(devKey) ||
                    devKey.toString() === "11111111111111111111111111111111" ? (
                    <></>
                ) : (
                    <WorkgroupDevice
                        key={devKey.toString()}
                        provider={provider}
                        deviceKey={devKey}
                        localDeviceData={localDeviceData}
                    />
                );
            })}
        </Accordion>
    );
}
