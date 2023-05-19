import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { Dropdown } from "react-bootstrap";
import { BsFillGearFill } from "react-icons/bs";
import { deviceInfo } from "../../hooks/meshInfo";
import { useProposer } from "../../hooks/proposer";
import { closeDevice } from "../../worknet/createWorkgroup";
import { toastAndSendProposedTransactions } from "../../worknet/displays";
import { getProvider, WORKNET_V1_PROGRAM_ID } from "../../worknet/worknet";
import * as anchor from "@project-serum/anchor";
import * as sol from "@solana/web3.js";

interface NodeSettingsProps {
    device: deviceInfo;
}
export function NodeSettings({ device }: NodeSettingsProps) {
    const wallet = useWallet();
    const { proposer } = useProposer();
    const { connection } = useConnection();
    const provider = getProvider(connection, wallet);
    // TODO: check that the wallet either has ownership of the device, or can modify the workgroup
    //       shows we need a way for a device that is in a workgroup, to store state that says, no, i'm connecting elsewhere
    //       which might be a way to allow devices to be on more than one workgroup, and pick what set to connect to.
    const walletConnected = wallet.connected;

    // TODO: need to enable "add-service" only for devices the user has rights
    //       which shows we need to add some kind of ownership contept

    return (
        <div>
            <Dropdown>
                <Dropdown.Toggle
                    size="sm"
                    variant="outline-secondary"
                    id="dropdown-node-settings"
                >
                    <BsFillGearFill />
                    <span className="ms-2">Settings</span>
                </Dropdown.Toggle>
                <Dropdown.Menu>
                    <Dropdown.Item
                        disabled={device.Info.Status == 0}
                        href="#/add-service"
                    >
                        Add service
                    </Dropdown.Item>
                    <Dropdown.Item
                        href="#/remove-device"
                        disabled={!walletConnected}
                        onClick={async () => {
                            console.log({ device: device });
                            if (!proposer || !device.Info.DeviceAuthority) {
                                return;
                            }

                            // eslint-disable-next-line no-unused-vars
                            const [devicePDA] =
                                await anchor.web3.PublicKey.findProgramAddress(
                                    [
                                        new sol.PublicKey(
                                            device.Info.DeviceAuthority.toString()
                                        ).toBuffer(),
                                    ],
                                    new anchor.web3.PublicKey(
                                        WORKNET_V1_PROGRAM_ID
                                    )
                                );
                            toastAndSendProposedTransactions(
                                "CloseDevice",
                                proposer,
                                provider.connection,
                                await closeDevice(
                                    provider,
                                    proposer,
                                    device.Info.WorkGroup,
                                    devicePDA
                                )
                            );
                        }}
                    >
                        Remove device
                    </Dropdown.Item>
                </Dropdown.Menu>
            </Dropdown>
        </div>
    );
}
