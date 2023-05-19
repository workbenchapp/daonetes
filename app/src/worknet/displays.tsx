import * as sol from "@solana/web3.js";
import { toast } from "react-toastify";
import { ProposalDetails, Proposer } from "./createWorkgroup";

export function solScanTxnURL(clusterURL: string, txn: string): string {
    return `https://solscan.io/tx/${txn}?cluster=custom&customUrl=${encodeURIComponent(
        clusterURL
    )}`;
}

export function transactionSuccess(
    connection: sol.Connection,
    proposalURL: string | undefined
) {
    const successFn = ({ data }: { data?: any }) => {
        return (
            <div>
                <div className="text-success">Success</div>
                <div className="d-flex justify-content-between">
                    <a
                        target="_blank"
                        rel="noreferrer"
                        href={solScanTxnURL(connection.rpcEndpoint, data || "")}
                    >
                        Transaction
                    </a>
                    {proposalURL && (
                        <a target="_blank" rel="noreferrer" href={proposalURL}>
                            Proposal
                        </a>
                    )}
                </div>
            </div>
        );
    };
    return successFn;
}

export function explorerCluster(rpcEndpoint: string): string {
    if (rpcEndpoint.includes("devnet")) return "devnet";
    if (rpcEndpoint.includes("mainnet-beta")) return "mainnet-beta";
    // TODO: Other RPCs 🙈
    return "custom";
}

export function transactionError(
    connection: sol.Connection,
    txn: sol.Transaction
) {
    const errorFn = ({ data }: { data?: any }) => {
        const inspectURL = `https://explorer.solana.com/tx/inspector?cluster=${explorerCluster(
            connection.rpcEndpoint
        )}&message=${encodeURIComponent(
            txn.serializeMessage().toString("base64")
        )}`;
        console.log("transaction error", { data, inspectURL });
        let msg = `Unknown error. ${data.logs
            .filter((l: string) => l.toLowerCase().includes("error"))
            .join(" ")}`;
        if (data.message.includes("0x23c")) {
            msg = "Too many proposals outstanding. Cancel some. CODE=0x23c";
        }
        if (data.message.includes("0x20b")) {
            msg =
                "Invalid state. Can't edit transactions in proposal. CODE=0x20b";
        }
        return (
            <div>
                <div>
                    {msg}
                    <div>
                        <a target="_blank" rel="noreferrer" href={inspectURL}>
                            Inspect Simulation
                        </a>
                    </div>
                </div>
            </div>
        );
    };
    return errorFn;
}

export async function toastAndSendProposedTransactions(
    innerInstructionName: string,
    proposer: Proposer,
    connection: sol.Connection,
    proposalDetails: ProposalDetails
): Promise<void> {
    const { txns } = proposalDetails;
    const serializeAndSend = (txn: sol.Transaction) => {
        return sol.sendAndConfirmRawTransaction(connection, txn.serialize(), {
            commitment: "confirmed",
        });
    };

    const toastParams = (
        msg: string,
        txn: sol.Transaction,
        proposalURL?: string
    ) => {
        return {
            pending: msg,
            success: {
                render: transactionSuccess(connection, proposalURL),
            },
            error: {
                render: transactionError(connection, txn),
            },
            autoClose: 10000,
            hideProgressBar: true,
        };
    };

    if (proposer.dao()) {
        const [propose, insert] = txns;
        if (!propose || !insert || !propose.signature || !insert.signature)
            throw `Didn't get signed DAO transactions expected from ${innerInstructionName}`;

        await toast.promise(
            serializeAndSend(propose),
            toastParams(
                `Creating new DAO proposal for ${innerInstructionName}`,
                propose,
                proposalDetails.url
            )
        );

        await toast.promise(
            serializeAndSend(insert),
            toastParams(
                `Inserting transaction for ${innerInstructionName} to proposal`,
                insert
            )
        );

        return;
    }

    if (!txns[0] || !txns[0].signature)
        throw `Didn't get signed transaction expected for ${innerInstructionName}`;

    await toast.promise(
        serializeAndSend(txns[0]),
        toastParams(`Running ${innerInstructionName}`, txns[0])
    );
}
