import type { NextPage } from "next";

import { Card, Row, Badge, Col } from "react-bootstrap";
import { OnBoardingWizard } from "../components/OnBoardingWizard";
import { useQuery } from "@tanstack/react-query";
import { getLocalhostDeviceInfo, getLocalhostEndpointInfo } from "../hooks/localhostDevice";
import { useWorkgroup } from "../hooks/selectedWorkgroup";
import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { getDevice, getProvider, getWorknetGroup } from "../worknet/worknet";
import { PublicKey } from "@solana/web3.js";
import { AnchorProvider } from "@project-serum/anchor";

type ComposeService = {
    Command: string;
    ExitCode: number;
    Health: string;
    ID: string;
    Name: string;
    Project: string;
    Service: string;
    State: string;
    Publishers: {
        Protocol: string;
        PublishedPort: number;
        TargetPort: number;
        URL: string;
    }[];
};
type DeployStateServicesProps = {
    key: string;
    deployState: any;
    deviceInfo: any;
};
function DeployStateServices({
    key,
    deployState,
    deviceInfo,
}: DeployStateServicesProps) {
    const { status: endpointStatus, data: endpointInfo /*, error */ } = useQuery(["localendpointinfo"], async () => {
        try {
            const response = await getLocalhostEndpointInfo();
            console.log("getLocalhostEndpointInfo 22", { response })
            return response
        } catch (error) {
            console.log(error)
            return { error }
        }
    });
    if (endpointStatus === "loading") {
        return (<div>loading...</div>)
    }
    if (endpointStatus === "error") {
        return (<div>loading...</div>)
    }

    const services = deployState.state as ComposeService[];
    if (services.length === 0) {
        // TODO: need to get any error messaged from the agent's attempt
        return <div>{key} has No services running</div>;
    }

    // TODO: The publish list is way too much information - need to make the service.Name a link to the right URL.
    // TODO: need to see if we can convert the 0.0.0.0 to a real IP or even better a DNS entry...

    return (
        <div>
            {services.map((service) => {
                return (
                    <div key={service.ID}>
                        <ul>
                            <li>
                                {service.Name}
                                <Badge className="fw-normal" bg="success">
                                    {service.State}
                                </Badge>
                            </li>
                            <ul>
                                {service.Publishers?.map((publish, index) => {
                                    if (publish.PublishedPort === 0) {
                                        return (
                                            <li key={service.ID + index}>
                                                {publish.TargetPort} /{" "}
                                                {publish.Protocol} not published
                                            </li>
                                        );
                                    }
                                    if (publish.URL.includes(":")) {
                                        // TODO: skipping IPv6
                                        return <></>
                                    }
                                    var destinationUrl = deviceInfo.Hostname + ":" + publish.PublishedPort
                                    // TODO: need to figure out local services - they're not in the enpoints API atm
                                    if (endpointInfo) {
                                        const endpointInfoMap = new Map(Object.entries(endpointInfo))
                                        const test = endpointInfoMap.get(service.Name + ":" + publish.PublishedPort)
                                        if (test) {
                                            destinationUrl = test
                                        }
                                    }
                                    return (
                                        <li key={service.ID + index}>
                                            <a href={"http://" + destinationUrl}>
                                                {destinationUrl} /{" "}
                                                {publish.Protocol} connects to{" "}
                                                {deviceInfo.Hostname}:{" "}
                                                {publish.TargetPort}
                                            </a>
                                        </li>
                                    );
                                })}
                            </ul>
                        </ul>
                    </div>
                );
            })}
        </div>
    );
}

type DeploymentInfoProps = {
    key: string;
    deployState: any;
    deviceData: any;
};
function DeploymentInfo({ key, deployState, deviceData }: DeploymentInfoProps) {
    return (
        <Col>
            <Card key={key}>
                <Card.Header>
                    <Badge className="fw-normal" bg="primary">
                        {deviceData.deviceInfo.Hostname}
                    </Badge>
                    {"   "} : {"   "}
                    {deployState.deployment.Name}
                    {/* <Image alt={deployState.spec.Name} src="image.png" /> */}
                </Card.Header>
                <Card.Body>
                    <Card.Text>
                        <DeployStateServices
                            key={key}
                            deployState={deployState}
                            deviceInfo={deviceData.deviceInfo}
                        />
                    </Card.Text>
                </Card.Body>
            </Card>
        </Col>
    );
}

type DeviceDeploymentsProps = {
    hostname: string;
};
function DeviceDeployments({ hostname }: DeviceDeploymentsProps) {
    const { status: deviceStatus, data: deviceData /*, error */ } = useQuery(
        ["agentinfo", hostname],
        async () => {
            try {
                const response = await getLocalhostDeviceInfo(hostname);
                console.log("getLocalhostDeviceInfo 22", { response });
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

    if (!deviceData.deviceInfo) {
        return <></>;
        //return <div>Failed to connect to any agents, please Manage your Workgroup</div>
    }
    const listOfDeployments = deviceData.deploy_state;
    if (!listOfDeployments) {
        return (
            <Card key={deviceData.deviceInfo.Hostname}>
                <Card.Body>
                    <Card.Title>
                        <Badge className="fw-normal" bg="primary">
                            {deviceData.deviceInfo.Hostname}
                        </Badge>
                        {"   "} : {"   "}
                        No services deployed on {deviceData.deviceInfo.Hostname}
                    </Card.Title>
                    <Card.Text>Maybe you should add some :)</Card.Text>
                </Card.Body>
            </Card>
        );
    }

    return (
        <>
            {Object.keys(listOfDeployments).map((key) => {
                const deployState = listOfDeployments[key];
                console.log(key, { deployState });
                return (
                    <DeploymentInfo
                        key={key}
                        deployState={deployState}
                        deviceData={deviceData}
                    />
                );
            })}
        </>
    );
}

type DeviceKeyDeploymentsProps = {
    provider: AnchorProvider;
    deviceKey: PublicKey;
};
function DeviceKeyDeployments({
    provider,
    deviceKey,
}: DeviceKeyDeploymentsProps) {
    const { status, data /*, error */ } = useQuery(
        ["device", deviceKey.toString()],
        async () => {
            return await getDevice(provider, deviceKey);
        }
    );

    if (status === "loading") return <div>loading...</div>;
    if (status === "error") return <></>;
    if (data.hostname === "") return <></>;
    return <DeviceDeployments hostname={data.hostname} />;
}

type ServicesListProps = {
    workgroupKey: PublicKey;
};
function ServicesList({ workgroupKey }: ServicesListProps) {
    const wallet = useWallet();
    const { connection } = useConnection();
    const provider = getProvider(connection, wallet);

    const {
        status: workgroupQueryStatus,
        data: group,
        //error,
    } = useQuery(["workgroup", workgroupKey.toString()], async () => {
        return getWorknetGroup(provider, workgroupKey);
    });

    if (workgroupQueryStatus === "loading") {
        return <div>loading...</div>;
    }

    return (
        <div>
            <h2>Services deployed to workgroup</h2>
            <div className="flex-1 w-full min-h-0">
                {/* <ButtonToolbar>
                    <Button>Add device</Button>
                    <Button>Add spec</Button>
                    <Button>Add deployment</Button>
                </ButtonToolbar> */}
                <Row xs={1} md={2} className="g-4">
                    {group?.devices.map((deviceKey) => {
                        return (
                            <DeviceKeyDeployments
                                key={deviceKey.toString()}
                                deviceKey={deviceKey}
                                provider={provider}
                            />
                        );
                    })}
                </Row>
            </div>
        </div>
    );
}

const Dashboard: NextPage = () => {
    const workgroupKey = useWorkgroup();

    if (!workgroupKey) {
        return (
            <OnBoardingWizard />
        )
    }

    if (!workgroupKey) {
        return (
            <div>
                new user onboarding wizard. connect wallet, select group owner,
                go. plus info on what it does.
            </div>
        );
    }

    return <ServicesList workgroupKey={workgroupKey} />;
};

export default Dashboard;
