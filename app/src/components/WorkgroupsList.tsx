
import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { useQuery } from "@tanstack/react-query";
import { Badge, Col, ListGroup, Nav, NavDropdown, Row } from "react-bootstrap";
import { getLocalhostDeviceInfo } from "../hooks/localhostDevice";
import { useWorkgroupUrl, useWorkgroup } from "../hooks/selectedWorkgroup";
import { useValidatorNetworkName } from "../hooks/validatorNetwork";
import { getProvider, getWorknets } from "../worknet/worknet";

interface GroupListItemProps {
    active: boolean;
    group: any; // TODO: its the anchor workgroup..
}
function GroupListItem({ active, group }: GroupListItemProps) {
    const wallet = useWallet();
    const url = useWorkgroupUrl(group.publicKey);
    const { status, data: deviceData } = useQuery(
        ["localdeviceinfo"],
        async () => {
            return await getLocalhostDeviceInfo();
        }
    );
    if (status === "loading") {
        return <div>loading...</div>;
    }

    // show different types
    // - personal (highlight the ones owned by the current wallet)
    // - multisig (maybe highlight the ones the current wallet is in)
    // - realms (again, aybe highlight the ones the wallet is in)
    const myGroup =
        group.account.groupAuthority.toString() ===
            wallet.publicKey?.toString() ? (
            <Badge className="fw-normal" bg="primary">
                Selected Wallet
            </Badge>
        ) : (
            <></>
        );
    // Some kind of "HEY THIS IS THE COMPUTER YOU'RE USING THE BROWSER ON...
    const myLocalDeviceGroup =
        deviceData &&
            deviceData.deviceInfo &&
            deviceData.deviceInfo.WorkGroup === group.publicKey.toString() ? (
            <Badge className="fw-normal" bg="success">
                Local device
            </Badge>
        ) : (
            <></>
        );
    return (
        <ListGroup.Item active={active} action href={"?" + url.toString()}>
            <Row>
                <Col md={true}>
                    {group.account.name} - {group.account.identifier}
                </Col>
                <Col md={4}>
                    {myGroup}
                    {myLocalDeviceGroup}
                </Col>
            </Row>
        </ListGroup.Item>
    );
}

export function WorkgroupsList() {
    const wallet = useWallet();
    const workgroup = useWorkgroup();
    const { connection } = useConnection();
    const provider = getProvider(connection, wallet);

    const { status, data: groups /*, error*/ } = useQuery(
        ["workgroups"],
        async () => {
            // TODO: should really load them all into the individual cache too..
            return getWorknets(provider);
        }
    );
    if (status === "loading") {
        return <div>loading...</div>;
    }
    return (
        <ListGroup>
            {groups?.map((group: any) => {
                const currentGroup = workgroup
                    ? group.publicKey.equals(workgroup)
                    : false;

                return (
                    <GroupListItem
                        active={currentGroup}
                        group={group}
                        key={group.publicKey.toString()}
                    />
                );
            })}
        </ListGroup>
    );
}

interface DaonetesWorkGroupItemProps {
    group: any; // TODO: argh!
}
function DaonetesWorkGroupItem({ group }: DaonetesWorkGroupItemProps) {
    const url = useWorkgroupUrl(group.publicKey);

    return (
        <Nav.Link>
            <NavDropdown.Item href={"?" + url.toString()}>
                {group.account.name} - {group.account.identifier}
            </NavDropdown.Item>
        </Nav.Link>
    );
}
// TODO: this should be the global selector of what workgroup we're looking at
// TODO: but should also help us make a new one..?
/*
SVEN - plan.
I want to support workgroups attached to
1. a plain user's account
2. a goki multisig
3. a governance realm.

in all 3 cases, whatever the user selects (or pastes) into the selector needs to be converted into the PDA that contains the worknet data
*/
export function DaonetesWorkGroup() {
    const wallet = useWallet();
    const workgroup = useWorkgroup();
    const { connection } = useConnection();
    const provider = getProvider(connection, wallet);
    const networkName = useValidatorNetworkName();
    const newWorkgroupURL = new URLSearchParams();
    newWorkgroupURL.set("network", networkName);

    const { status, data: groups /*, error*/ } = useQuery(
        ["workgroups"],
        async () => {
            // TODO: should really load them all into the individual cache too..
            return getWorknets(provider);
        }
    );
    if (status === "loading") {
        return <div>loading...</div>;
    }

    const selectedGroup = groups?.find((group: any) => {
        if (workgroup && group.publicKey.equals(workgroup)) {
            return true;
        }
        return false;
    });
    const owner = selectedGroup
        ? selectedGroup.account.name + " - " + selectedGroup.account.identifier
        : "Select Workgroup";

    //return <NavItem>ARGH! {worknetBaseWalletPubKey.toString()}</NavItem>;
    //             <DaonetesWorkGroupItem name="Test GOKI wallet" walletKey={testGokiWallet.toString()} />

    return (
        <NavDropdown title={owner} id="navbarScrollingDropdown" drop="up">
            {groups?.map((group: any) => {
                return (
                    <DaonetesWorkGroupItem
                        group={group}
                        key={group.publicKey.toString()}
                    />
                );
            })}

            <NavDropdown.Divider />
            <NavDropdown.Item href={"?" + newWorkgroupURL.toString()}>
                + Create New
            </NavDropdown.Item>
        </NavDropdown>
    );
}

export default WorkgroupsList;
