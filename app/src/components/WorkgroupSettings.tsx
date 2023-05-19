import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { PublicKey } from "@solana/web3.js";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { Button, Dropdown, Modal } from "react-bootstrap";
import { BsFillGearFill } from "react-icons/bs";
import { toast } from "react-toastify";
import { useProposer } from "../hooks/proposer";
import { closeWorkGroup } from "../worknet/createWorkgroup";
import { toastAndSendProposedTransactions } from "../worknet/displays";
import { getProvider, getWorknetGroup } from "../worknet/worknet";

export function WorkgroupSettings({
    workgroupKey,
}: {
    workgroupKey: PublicKey;
}) {
    const wallet = useWallet();
    const { connection } = useConnection();
    const provider = getProvider(connection, wallet);
    const { proposer } = useProposer();
    const { data: group } = useQuery(
        ["workgroup", workgroupKey.toString()],
        () => getWorknetGroup(provider, workgroupKey)
    );

    const [closingWorkgroup, setClosingWorkgroup] = useState(false);
    const [pastedWorkgroupKey, setPastedWorkgroupKey] = useState("");

    return (
        <div>
            <Dropdown>
                <Dropdown.Toggle variant="secondary" id="dropdown-basic">
                    <BsFillGearFill />
                    <span className="ms-2">Settings</span>
                </Dropdown.Toggle>
                <Dropdown.Menu>
                    <Dropdown.Item
                        onClick={() => setClosingWorkgroup(true)}
                        href="#/close-workgroup"
                    >
                        Close Workgroup
                    </Dropdown.Item>
                </Dropdown.Menu>
            </Dropdown>
            <Modal
                onHide={() => setClosingWorkgroup(false)}
                show={closingWorkgroup}
            >
                <Modal.Header closeButton>
                    <Modal.Title>Close Workgroup?</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <p>
                        This operation will close the workgroup and all
                        associated accounts and refund the SOL to the group's
                        authority.
                    </p>
                    <p>
                        <span className="text-danger">
                            {" "}
                            This is a destructive action and cannot be undone.
                        </span>
                    </p>
                    <p>
                        Please paste in the public key of the group below to
                        confirm.
                    </p>
                    <input
                        className="form-control text-monospace"
                        onChange={(e) => setPastedWorkgroupKey(e.target.value)}
                        type="text"
                    />
                </Modal.Body>
                <Modal.Footer>
                    <Button
                        onClick={async () => {
                            if (
                                workgroupKey.toString() === pastedWorkgroupKey
                            ) {
                                console.log({ proposer, group });
                                if (!proposer || !group)
                                    throw "Proposer or work group missing";

                                await toastAndSendProposedTransactions(
                                    `CloseWorkGroup`,
                                    proposer,
                                    provider.connection,
                                    await closeWorkGroup(
                                        provider,
                                        proposer,
                                        workgroupKey
                                    )
                                );
                            } else {
                                toast.error("Workgroup keys do not match");
                            }
                        }}
                        variant="danger"
                    >
                        Close Workgroup
                    </Button>
                </Modal.Footer>
            </Modal>
        </div>
    );
}
