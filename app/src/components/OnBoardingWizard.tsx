import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { useEffect, useState } from "react";
import { Button, Form, Stack } from "react-bootstrap";
import Split from "react-split";
import { getLicenseMint, getProvider } from "../worknet/worknet";
import WorkgroupsList from "./WorkgroupsList";
import {
    AccountInfo,
    ParsedAccountData,
    PublicKey,
    RpcResponseAndContext,
} from "@solana/web3.js";
import { createWorkgroup, RealmsProposer } from "../worknet/createWorkgroup";
import { generateMnemonic } from "bip39";
import Image from "next/image";
import { useProposer } from "../hooks/proposer";
import { toastAndSendProposedTransactions } from "../worknet/displays";

// TODO: These enums are 100% used, no idea
// why ESLint is complaining about this.
//
// eslint-disable-next-line no-unused-vars
enum WorkgroupCreateMode {
    // eslint-disable-next-line no-unused-vars
    DAO = "dao",
    // eslint-disable-next-line no-unused-vars
    Solo = "solo",
}

export function OnBoardingWizard() {
    const wallet = useWallet();
    const [groupID, setGroupID] = useState(
        generateMnemonic().split(" ").slice(0, 3).join("-")
    );
    const [groupName, setGroupName] = useState("default");
    const { proposer, setProposer } = useProposer();
    const { connection } = useConnection();
    const provider = getProvider(connection, wallet);
    const [daoKey, setDAOKey] = useState("");
    const [workgroupCreateMode, setWorkgroupCreateMode] = useState("");
    const [licenseTokens, setLicenseTokens] =
        useState<
            RpcResponseAndContext<
                { pubkey: PublicKey; account: AccountInfo<ParsedAccountData> }[]
            >
        >();

    const licenseMint = getLicenseMint(provider);

    useEffect(() => {
        const checkTokens = async () => {
            if (!proposer) return;
            const depositingLicenseTokensResp =
                await connection.getParsedTokenAccountsByOwner(
                    await proposer.payer(),
                    {
                        mint: licenseMint,
                    }
                );
            if (depositingLicenseTokensResp === undefined)
                throw "Error calling RPC to check for workgroup license tokens";
            setLicenseTokens(depositingLicenseTokensResp);
        };
        checkTokens();
    }, [licenseMint, proposer, connection]);

    var helpMessage = (
        <small className="text-secondary">
            This will create a cluster controlled by only one private key. In
            order to run workloads, you will only need to sign transactions with
            your wallet.
        </small>
    );
    var tokenCountInfo = <></>;
    let tokenCount = 0;

    if (licenseTokens) {
        const parsed = licenseTokens.value[0]?.account.data.parsed;
        tokenCount = parsed ? parsed.info.tokenAmount.uiAmount : 0;
        if (tokenCount === 0) {
            tokenCountInfo = (
                <div>
                    <div>
                        <strong>No license tokens found.</strong> In order to
                        use DAOnetes, you need license token(s) in your wallet.
                        Please{" "}
                        <a
                            rel="noreferrer"
                            href="https://twitter.com/daonetes"
                            target="_blank"
                        >
                            request some
                        </a>
                        .
                    </div>
                </div>
            );
        } else {
            tokenCountInfo = (
                <span className="text-success">{`This ${
                    daoKey === "" ? "wallet" : "governance"
                } has ${tokenCount} DAOnetes license token(s).`}</span>
            );
        }
    }

    const handleCreateWorkGroup = async () => {
        if (!licenseTokens || !proposer || !licenseTokens.value[0]?.pubkey)
            return;

        toastAndSendProposedTransactions(
            "CreateWorkGroup",
            proposer,
            provider.connection,
            await createWorkgroup(
                provider,
                proposer,
                groupID,
                groupName,
                licenseTokens.value[0]?.pubkey
            )
        );

        // TODO: redirect to newly created workgroup page
    };

    if (!wallet.publicKey) {
        return (
            <div className="d-flex flex-column">
                <div className="d-flex justify-content-center">
                    <p className="text-secondary">
                        No wallet detected. Please connect your wallet to use
                        DAOnetes.
                    </p>
                </div>
                <div className="d-flex justify-content-center">
                    <Image
                        alt="DAOnetes logo, a dragonfly on turquoise crystalline background"
                        src="/icon.svg"
                        width={512}
                        height={512}
                    />
                </div>
            </div>
        );
    }

    const workgroupTypeOptions = [
        { value: WorkgroupCreateMode.Solo, name: "Solo User" },
        { value: WorkgroupCreateMode.DAO, name: "Multi User (DAO)" },
    ];

    return (
        <div>
            <Split
                sizes={[50, 50]}
                direction="horizontal"
                className="split vh-100"
                gutterSize={0}
            >
                <div className="m-2 almost-vh-100">
                    <Stack className="flex-fill almost-vh-100">
                        <h2>Create Workgroup</h2>
                        <Form>
                            <Form.Label>
                                <strong>Group Type</strong>
                            </Form.Label>
                            <Form.Select
                                as="select"
                                className="cursor-pointer"
                                value={workgroupCreateMode}
                                onChange={(e) => {
                                    setWorkgroupCreateMode(e.target.value);
                                }}
                            >
                                {workgroupTypeOptions.map((option) => (
                                    <option
                                        key={option.name}
                                        value={option.value}
                                    >
                                        {option.name}
                                    </option>
                                ))}
                            </Form.Select>
                        </Form>
                        {workgroupCreateMode === WorkgroupCreateMode.DAO ? (
                            <div>
                                <small className="text-secondary">
                                    This will create a cluster controlled by a{" "}
                                    <a
                                        className="link-secondary"
                                        href="https://realms.today"
                                        target="_blank"
                                        rel="noreferrer"
                                    >
                                        Realms DAO
                                    </a>
                                    . In order to run workloads, you will
                                    propose transactions to the DAO, which will
                                    vote and approve them to deploy.
                                </small>
                                <hr />
                                <div>
                                    <Form>
                                        <Form.Group>
                                            <Form.Label className="fw-bold">
                                                Workgroup ID
                                            </Form.Label>
                                            <small className="ms-2 text-secondary">
                                                A globally unique identifier for
                                                your group.
                                            </small>
                                            <Form.Control
                                                value={groupID}
                                                onChange={(e) =>
                                                    setGroupID(e.target.value)
                                                }
                                            />
                                            <Form.Label className="mt-1 fw-bold">
                                                Workgroup Name
                                            </Form.Label>
                                            <small className="ms-2 text-secondary">
                                                A human readable name for the
                                                group.
                                            </small>
                                            <Form.Control
                                                value={groupName}
                                                onChange={(e) =>
                                                    setGroupName(e.target.value)
                                                }
                                            />
                                            <div className="mt-2 d-flex justify-content-center">
                                                <Image
                                                    alt="Directions how to get the governance ID. Click Settings on the Realms main page, then get the governance ID from the governances section."
                                                    src="/realms_instructions.jpg"
                                                    height={540}
                                                    width={525}
                                                />
                                            </div>
                                            <Form.Label className="mt-2 fw-bold">
                                                Governance Pubkey
                                            </Form.Label>{" "}
                                            <small className="ms-2 text-secondary">
                                                The SPL governance that will
                                                control this group.
                                            </small>
                                            <Form.Control
                                                placeholder="Paste governance key here"
                                                className="font-monospace"
                                                onClick={() => {
                                                    setDAOKey("");
                                                    setProposer(undefined);
                                                }}
                                                onPaste={(e) => {
                                                    setLicenseTokens(undefined);
                                                    const governanceID =
                                                        e.clipboardData.getData(
                                                            "text"
                                                        );
                                                    setDAOKey(governanceID);
                                                    setProposer(
                                                        new RealmsProposer(
                                                            new PublicKey(
                                                                governanceID
                                                            ),
                                                            provider
                                                        )
                                                    );
                                                }}
                                            />
                                        </Form.Group>
                                        <div className="d-flex align-items-center justify-content-between">
                                            <div className="ms-2">
                                                {daoKey !== "" &&
                                                    licenseTokens &&
                                                    tokenCountInfo}
                                            </div>
                                            <div>
                                                {daoKey !== "" &&
                                                    tokenCount > 0 && (
                                                        <Button
                                                            onClick={
                                                                handleCreateWorkGroup
                                                            }
                                                            className="m-2"
                                                        >
                                                            Propose Create
                                                        </Button>
                                                    )}
                                            </div>
                                        </div>
                                    </Form>
                                </div>
                            </div>
                        ) : (
                            <div>
                                <div>{helpMessage}</div>
                                <hr />
                                <Form>
                                    <Form.Group>
                                        <Form.Label className="fw-bold">
                                            Workgroup ID
                                        </Form.Label>
                                        <small className="ms-2 text-secondary">
                                            A globally unique identifier for
                                            your group.
                                        </small>
                                        <Form.Control
                                            value={groupID}
                                            onChange={(e) =>
                                                setGroupID(e.target.value)
                                            }
                                        />
                                        <Form.Label className="mt-1 fw-bold">
                                            Workgroup Name
                                        </Form.Label>
                                        <small className="ms-2 text-secondary">
                                            A human readable name for the group.
                                        </small>
                                        <Form.Control
                                            value={groupName}
                                            onChange={(e) =>
                                                setGroupName(e.target.value)
                                            }
                                        />
                                    </Form.Group>
                                    <div className="d-flex align-items-center justify-content-between">
                                        <div>{tokenCountInfo}</div>
                                        {tokenCount > 0 && (
                                            <Button
                                                onClick={handleCreateWorkGroup}
                                                className="m-2"
                                            >
                                                Create
                                            </Button>
                                        )}
                                    </div>
                                </Form>
                            </div>
                        )}
                    </Stack>
                </div>
                <div>
                    <small className="text-secondary">or</small>
                    <h2>Select Workgroup</h2>
                    <div className="m-2">
                        <WorkgroupsList />
                    </div>
                </div>
            </Split>
        </div>
    );
}
