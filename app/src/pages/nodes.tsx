import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { useQuery } from "@tanstack/react-query";
import type { NextPage } from "next";
import { Button } from "react-bootstrap";
import { getLocalhostDeviceInfo } from "../hooks/localhostDevice";
import { deviceInfo, getMeshInfo } from "../hooks/meshInfo";
import { useProposer } from "../hooks/proposer";
import { useWorkgroup, getWorkgroupUrl } from "../hooks/selectedWorkgroup";
import { RealmsProposer, registerDevice } from "../worknet/createWorkgroup";
import { toastAndSendProposedTransactions } from "../worknet/displays";
import { realmsGovernanceFromWorkgroup } from "../worknet/governanceHelpers";
import { getProvider, getWorknetGroup } from "../worknet/worknet";
import {
    createColumnHelper,
    flexRender,
    getCoreRowModel,
    useReactTable,
} from '@tanstack/react-table';
import { ReactNode } from "react";
import humanizeDuration from "humanize-duration";
import { NodeSettings } from "../components/node/NodeSettings";
import { Waitlist } from "../components/Waitlist";
import { PublicKeyView } from "../components/PublicKeyView";
import { DeviceAdd } from "../components/DeviceAdd";
import { useValidatorNetworkName } from "../hooks/validatorNetwork";
import { useRouter } from "next/router";

interface RegisterDeviceButtonProps {
    deviceKey: string;
}
function RegisterDeviceButton({ deviceKey }: RegisterDeviceButtonProps) {
    const workgroupKey = useWorkgroup();
    if (!workgroupKey) {
        return (
            <Button disabled>Register Device (select workgroup first)</Button>
        )
    }
    return (<RegisterDeviceButtonActive deviceKey={deviceKey} />)

}
function RegisterDeviceButtonActive({ deviceKey }: RegisterDeviceButtonProps) {
    const wallet = useWallet();
    const { connection } = useConnection();
    const provider = getProvider(connection, wallet);
    const { proposer, setProposer } = useProposer();

    const workgroupKey = useWorkgroup();

    const { data: group } = useQuery(
        ["workgroup", workgroupKey?.toString()],
        () => {
            if (!workgroupKey) {
                throw "TODO: workgroupKey not defined"
            }
            return getWorknetGroup(provider, workgroupKey);
        }
    );

    const { error: govError } = useQuery(
        [
            "governance",
            group?.chain?.PubKey.toString(),
            wallet.publicKey?.toString(),
        ],
        () => realmsGovernanceFromWorkgroup(provider, group),
        {
            onSuccess: (governance) => {
                if (!governance || !wallet || !wallet.publicKey || !group)
                    return;
                if (!wallet.publicKey.equals(group.groupAuthority)) {
                    setProposer(new RealmsProposer(governance, provider));
                }
            },
        }
    );
    console.log({ govError });
    if (!workgroupKey) {
        return (<></>)
        //throw "Workgroup key should not be null when adding device";
    }
    return <>
        <Button
            disabled={!workgroupKey}
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

            }}
        >
            {proposer?.dao() ? "Propose " : ""}
            Register Device
        </Button>
    </>
}

interface ServicesInfoProps {
    device: deviceInfo;
}
function ServicesInfo({ device }: ServicesInfoProps) {
    var serviceMap = device.LocalProxyListeners;
    if (!(serviceMap instanceof Map)) {
        // TODO: Sven doesn't understand why the type interface says its a map, it is a map, but this function seems to think its not...
        serviceMap = new Map(Object.entries(serviceMap));
    }
    var keys: string[] = [];
    if (serviceMap) {
        serviceMap.forEach((value, key) => {
            keys.push(key)//meshInfoData.ProxyDevices[pk])
        })
    }

    return (
        <table
            key={device.WireguardPeerKey}
            className="node-service-list-table"
        >
            <tbody>
                {
                    keys.map((address): ReactNode => {
                        var url = device.Info.Hostname + ".dmesh:" + address.split(":")[1]
                        var name = serviceMap.get(address)
                        if (name === device.Info.Hostname + "-deviceAPI") {
                            // Skip the device API
                            return;
                        }
                        return (<tr key={address}>
                            <td>
                                <a target="_blank" rel="noreferrer" href={"http://" + url}>{name}</a>
                            </td>
                        </tr>);
                    })
                }
            </tbody>
        </table>
    )
}

interface VersionInfoProps {
    device: deviceInfo;
}
function VersionInfo({ device }: VersionInfoProps) {
    const { status: deviceStatus, data: deviceData, error } =
        useQuery(["agentinfo", device.Info.Hostname], async () => {
            return await getLocalhostDeviceInfo(device.Info.Hostname);
        });

    if (deviceStatus === "loading") {
        return (<div className="small">loading...</div>);
    }
    if (deviceStatus === "error") {
        return (<div className="small">Error: {JSON.stringify(error)}</div>);
    }
    if (deviceData === undefined) {
        return (<div className="small">unknown</div>)
    }
    var git = deviceData.VersionRevision as string || "";
    if (git.endsWith("-modified")) {
        git = git.substring(0, 8) + " Modified"
    } else {
        git = git.substring(0, 8)
    }
    var date = "";
    if (deviceData.VersionDate) {
        var versionDate = new Date(deviceData.VersionDate)
        date = versionDate.toLocaleDateString("en-UK", {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        })
    }
    return (
        <div className="small">
            <div>{deviceData.Version}</div>
            <span>{date}</span>
            <div>{git}</div>
        </div>
    )
}

interface MeshStatusProps {
    device: deviceInfo;
}
function MeshStatus({ device }: MeshStatusProps) {
    const { status: meshInfoStatus, data: meshInfoData, error: meshInfoError } =
        useQuery(["meshInfo", ""], async () => {
            return await getMeshInfo();
        });

    if (meshInfoStatus === "loading") {
        return <div className="small">meshInfoStatus loading...</div>;
    }
    if (meshInfoStatus === "error") {
        return <div className="small">ERROR: {JSON.stringify(meshInfoError)}</div>;
    }


    var status = ""
    // TODO: replace forEach with find()
    meshInfoData?.Peers?.forEach((value) => {
        if (value.AllowedIPs[0]?.IP == device.WireguardAddress) {
            var lastSeen = Date.parse(value.LastHandshakeTime)
            const timeSince = Date.now() - lastSeen
            if (timeSince < 10 * 60 * 1000) { // <5mins?
                status = humanizeDuration(
                    timeSince,
                    {
                        // TODO: language:
                        largest: 1,
                        round: true,
                    })
            } else {
                var lastDate = new Date(lastSeen)
                status = lastDate.toLocaleDateString("en-UK", {
                    weekday: 'long',
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric'
                })
            }

        }
    })

    return (
        <div className="small">{status}</div>
    )
}

const columnHelper = createColumnHelper<deviceInfo>()
const columnMapping = [
    columnHelper.accessor(row => row.Info.Hostname, {
        id: 'hostname',
        cell: info => <b>{info.getValue() + ".dmesh"}</b>,  // TODO: add a popup that shows detailed info
        header: () => <span>Host Name</span>,
        //footer: info => info.column.id,
    }),
    columnHelper.accessor(row => row.ProxyAddress, {
        id: 'localaddress',
        cell: info => <code>{info.getValue()}</code>,
        header: () => <span>Local Address</span>,
        //footer: info => info.column.id,
    }),
    columnHelper.accessor(row => row, {
        id: 'services',
        cell: info => <ServicesInfo device={info.getValue()} />,
        header: () => <span>Mapped Services</span>,
        //footer: info => info.column.id,
    }),
    columnHelper.accessor(row => row, {
        id: 'version',
        cell: info => <VersionInfo device={info.getValue()} />,
        header: () => <span>Agent version</span>,
        //footer: info => info.column.id,
    }),
    columnHelper.accessor(row => row, {
        id: 'meshstatus',
        cell: info => <MeshStatus device={info.getValue()} />,
        header: () => <span>Last keepalive</span>,
        //footer: info => info.column.id,
    }),
    columnHelper.accessor(row => row, {
        id: 'Action',
        cell: info => <NodeSettings device={info.getValue()} />,
        header: () => <span>Actions</span>,
        //footer: info => info.column.id,
        footer: () => <DeviceAdd />,
    }),
]

const Nodes: NextPage = () => {
    const router = useRouter();

    const networkName = useValidatorNetworkName();
    const { status: deviceStatus, data: localDeviceData, error: localDeviceError } =
        useQuery(["agentinfo", "localhost"], async () => {
            return await getLocalhostDeviceInfo();
        });
    const { status: meshInfoStatus, data: meshInfoData, error: meshInfoError } =
        useQuery(["meshInfo", ""], async () => {
            return await getMeshInfo();
        });
    // Seriously, I can't figure out why what `npm run build` wants, and what my browser is willing to not crash on is, so.. this fucking works
    var devices: deviceInfo[] = [];
    if (meshInfoData?.ProxyDevices) {
        // TODO: again, passing a map seems to break on my browser
        const deviceList = new Map(Object.entries(meshInfoData?.ProxyDevices));

        // my browser accecpts this
        // for (let key in meshInfoData.ProxyDevices) {
        //     var pk = new PublicKey(key)
        //     devices.push(meshInfoData.ProxyDevices[pk])
        // }
        // vscode wants this (but the browser hates it)
        deviceList.forEach((value) => {
            devices.push(value);
        })
    }
    const table = useReactTable({
        data: devices,
        columns: columnMapping,
        getCoreRowModel: getCoreRowModel(),
    })
    const workgroupKey = useWorkgroup();

    if (deviceStatus === "loading") {
        return <div>loading...</div>;
    }
    if (deviceStatus === "error") {
        return <div>
            <h2>No DAOnetes agent running on your localhost computer</h2>
            <div>You my need to Download&Install the agent from (INSERT URL HERE) or start it.</div>
            <Waitlist deviceKey="" />
            <div>Debug details: {JSON.stringify(localDeviceError)}</div>
        </div>;
    }

    // lofalDeviceData.deviceInfo.Status === 0
    if (localDeviceData.deviceInfo.DeviceAuthority === "11111111111111111111111111111111") {
        return <>
            <Waitlist deviceKey={localDeviceData.deviceWallet} />
            Your local computer {localDeviceData.deviceWallet} is not yet registered.
            <br />
            This requires either:
            <ul>
                <li>
                    a funded onchain Solana Browser wallet <b>set to using devnet</b>, or
                </li>
                <li>
                    that you send your computer key to someone that is managing a workgroup you want to join, or
                </li>
                <li>
                    if you have admin rights to an existing Workgroup, use <RegisterDeviceButton deviceKey={localDeviceData.deviceWallet} />
                </li>
            </ul>
        </>
    }
    if (localDeviceData.deviceInfo.WorkGroup === "11111111111111111111111111111111") {
        return <>
            Device {localDeviceData.deviceWallet} not in a Workgroup
            <Waitlist deviceKey={localDeviceData.deviceWallet} />
            (this requires a funded onchain Browser wallet)

            OR, send your key to someone that is managing a workgroup you want to join

            OR, if you have admin rights to an existing Workgroup, use <RegisterDeviceButton deviceKey={localDeviceData.deviceWallet} />
        </>
    }
    if (!workgroupKey || workgroupKey.toString() != localDeviceData.deviceInfo.WorkGroup) {
        // TODO: tomfollery so we can make the add button work..
        const newUrl = getWorkgroupUrl(networkName, localDeviceData.deviceInfo.WorkGroup)
        router.push("?" + newUrl)
    }

    if (meshInfoStatus === "loading") {
        return <div>meshInfoStatus loading...</div>;
    }
    if (meshInfoStatus === "error") {
        return <div>
            <h3>Error getting Mesh info from your local DAOnetes agent</h3>
            <div>
                Please refresh your browser in a few seconds
            </div>
            <div>Debug details: {JSON.stringify(meshInfoError)}</div>
        </div>;
    }

    return (
        <>
            <p>
                {devices.length} Devices in the Workgroup: <PublicKeyView publicKey={localDeviceData?.deviceInfo?.WorkGroup} />
            </p>
            <table className="node-list-table">
                <thead>
                    {table.getHeaderGroups().map(headerGroup => (
                        <tr key={headerGroup.id}>
                            {headerGroup.headers.map(header => (
                                <th key={header.id}>
                                    {header.isPlaceholder
                                        ? null
                                        : flexRender(
                                            header.column.columnDef.header,
                                            header.getContext()
                                        )}
                                </th>
                            ))}
                        </tr>
                    ))}
                </thead>
                <tbody>
                    {table.getRowModel().rows.map(row => (
                        <tr key={row.id}>
                            {row.getVisibleCells().map(cell => (
                                <td key={cell.id}>
                                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                </td>
                            ))}
                        </tr>
                    ))}
                </tbody>
                <tfoot>
                    {table.getFooterGroups().map(footerGroup => (
                        <tr key={footerGroup.id}>
                            {footerGroup.headers.map(header => (
                                <th key={header.id}>
                                    {header.isPlaceholder
                                        ? null
                                        : flexRender(
                                            header.column.columnDef.footer,
                                            header.getContext()
                                        )}
                                </th>
                            ))}
                        </tr>
                    ))}
                </tfoot>
            </table>
            <p>
            </p>
        </>
    )
};

export default Nodes;