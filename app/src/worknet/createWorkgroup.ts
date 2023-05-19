import {
    LAMPORTS_PER_SOL,
    PublicKey,
    TransactionInstruction,
} from "@solana/web3.js";
import * as sol from "@solana/web3.js";

//import * as web3 from "@solana/web3.js";
import * as anchor from "@project-serum/anchor";
import { AnchorProvider } from "@project-serum/anchor";

const { TOKEN_PROGRAM_ID } = require("@solana/spl-token");

const { SystemProgram, SYSVAR_RENT_PUBKEY } = anchor.web3;
import {
    VoteType,
    withCreateProposal,
    PROGRAM_VERSION_V2 as REALMS_PROGRAM_VERSION_V2,
    getTokenOwnerRecordAddress,
    getGovernance,
    getAllProposals,
    getNativeTreasuryAddress,
    withAddSignatory,
    withInsertTransaction,
    createInstructionData,
    withSignOffProposal,
    withCreateTokenOwnerRecord,
} from "@solana/spl-governance";

import { getLicenseMint, getWorknetProgram } from "./worknet";
import { explorerCluster } from "./displays";
import { getAssociatedTokenAddress } from "@solana/spl-token";

export interface ProposalDetails {
    txns: sol.Transaction[];
    url: string | undefined;
}

export interface Proposer {
    // TODO: Fix interface unused vars eslint weirdness

    // treasury returns the paying wallet. This is called a "treasury"
    // in SPL governance, "base wallet" in Goki, etc.
    payer(): Promise<sol.PublicKey>;

    // dao returns true if the proposer will need to go through
    // a DAO to get final approval (all proposals pass automatically
    // on local wallet)
    dao(): boolean;

    // proposeTxn wraps a given transaction (e.g., create workgroup)
    // in the relevant instructions for the governance system.
    proposeTxn(
        // eslint-disable-next-line no-unused-vars
        proposedInstructions: sol.TransactionInstruction[],
        // eslint-disable-next-line no-unused-vars
        proposalName?: string,
        // eslint-disable-next-line no-unused-vars
        proposalDetails?: string
    ): Promise<ProposalDetails>;
}

// Not really "proposing" anything, but exists to conform the method for
// governance proposing to the same as using your local keypair.
export class LocalWalletProposer implements Proposer {
    provider: AnchorProvider;

    constructor(provider: AnchorProvider) {
        this.provider = provider;
    }

    dao(): boolean {
        return false;
    }

    async payer(): Promise<sol.PublicKey> {
        return this.provider.publicKey;
    }

    async proposeTxn(
        proposedInstructions: sol.TransactionInstruction[]
    ): Promise<ProposalDetails> {
        let txn = new sol.Transaction();
        const { blockhash } =
            await this.provider.connection.getLatestBlockhash();
        txn.recentBlockhash = blockhash;
        txn.feePayer = this.provider.wallet.publicKey;
        proposedInstructions.forEach((i) => txn.add(i));
        return { txns: [txn], url: undefined };
    }
}

export class RealmsProposer implements Proposer {
    realmsProgramID = new PublicKey(
        "GovER5Lthms3bLBqWub97yVrMmEogzX7xNjdXpPPCVZw"
    );
    governanceKey: sol.PublicKey;
    provider: AnchorProvider;
    treasury: sol.PublicKey | undefined;

    constructor(
        governanceKey: sol.PublicKey,
        provider: AnchorProvider,
        customRealmsProgramID?: sol.PublicKey
    ) {
        if (customRealmsProgramID) this.realmsProgramID = customRealmsProgramID;
        this.governanceKey = governanceKey;
        this.provider = provider;
    }

    dao(): boolean {
        return true;
    }

    async payer(): Promise<sol.PublicKey> {
        if (this.treasury) return this.treasury;
        this.treasury = await getNativeTreasuryAddress(
            this.realmsProgramID,
            this.governanceKey
        );
        return this.treasury;
    }

    async proposeTxn(
        proposedInstructions: sol.TransactionInstruction[],
        proposalName: string,
        proposalDetails: string
    ): Promise<ProposalDetails> {
        let returnedTxns = [];

        let proposeAndAddSignatoryInstructions: TransactionInstruction[] = [];

        const voteType = VoteType.SINGLE_CHOICE;
        const options = ["Approve"];
        const useDenyOption = true;

        const governance = await getGovernance(
            this.provider.connection,
            this.governanceKey
        );

        const tokenOwnerRecord = await getTokenOwnerRecordAddress(
            this.realmsProgramID,
            governance.account.realm,
            governance.account.governedAccount,
            this.provider.wallet.publicKey
        );

        const tokenOwnerRecordInfo =
            await this.provider.connection.getAccountInfo(tokenOwnerRecord);

        if (!tokenOwnerRecordInfo) {
            await withCreateTokenOwnerRecord(
                proposeAndAddSignatoryInstructions,
                this.realmsProgramID,
                REALMS_PROGRAM_VERSION_V2,
                governance.account.realm,
                this.provider.wallet.publicKey,
                governance.account.governedAccount,
                this.provider.wallet.publicKey
            );
        }

        const allProposals = await getAllProposals(
            this.provider.connection,
            this.realmsProgramID,
            governance.account.realm
        );

        // Weird format of allProposals return, seems to return a Proposal[]
        // for each governance, so we iterate through and check to get the
        // length of the one that matters to us (e.g., council)
        let proposalIndex = 0;
        allProposals.forEach((proposalList) => {
            if (proposalList.length === 0) return;
            if (!proposalList[0]?.account.governance.equals(governance.pubkey))
                return;
            proposalIndex = proposalList.length;
        });

        const proposalAddress = await withCreateProposal(
            proposeAndAddSignatoryInstructions,
            this.realmsProgramID,
            REALMS_PROGRAM_VERSION_V2,
            governance.account.realm,
            governance.pubkey,
            tokenOwnerRecord,
            proposalName,
            proposalDetails,
            governance.account.governedAccount,
            this.provider.publicKey,
            proposalIndex,
            voteType,
            options,
            useDenyOption,
            this.provider.publicKey
        );

        // Add the proposal creator as the default signatory
        const signatoryRecord = await withAddSignatory(
            proposeAndAddSignatoryInstructions,
            this.realmsProgramID,
            REALMS_PROGRAM_VERSION_V2,
            proposalAddress,
            tokenOwnerRecord,
            this.provider.publicKey,
            this.provider.publicKey,
            this.provider.publicKey
        );

        console.log({
            proposeAndAddSignatoryInstructions,
            accounts: {
                realm: governance.account.realm.toString(),
                governance: governance.pubkey.toString(),
                tokenOwnerRecord: tokenOwnerRecord.toString(),
                governedAccount: governance.account.governedAccount.toString(),
                proposal: proposalAddress.toString(),
            },
            allProposals,
            proposalSeeds: {
                governance: governance.pubkey.toString(),
                governingTokenMint:
                    governance.account.governedAccount.toString(),
                proposalIndex: allProposals.length,
            },
            signatoryRecord: signatoryRecord.toString(),
        });

        let createProposalTxn = new sol.Transaction();
        proposeAndAddSignatoryInstructions.forEach((i) =>
            createProposalTxn.add(i)
        );
        returnedTxns.push(createProposalTxn);

        /*
        // dump txn for simulation
        createProposalTxn.recentBlockhash = (
            await this.provider.connection.getLatestBlockhash()
        ).blockhash;
        createProposalTxn.feePayer = this.provider.wallet.publicKey;
        console.log(
            (await this.provider.wallet.signTransaction(createProposalTxn))
                .serializeMessage()
                .toString("base64")
        );
        */

        let insertTxnInstructions: TransactionInstruction[] = [];

        await Promise.all(
            proposedInstructions.map((inst, i) =>
                withInsertTransaction(
                    insertTxnInstructions,
                    this.realmsProgramID,
                    REALMS_PROGRAM_VERSION_V2,
                    governance.pubkey,
                    proposalAddress,
                    tokenOwnerRecord,
                    this.provider.wallet.publicKey,
                    i,
                    0,
                    0,
                    [createInstructionData(inst)],
                    this.provider.publicKey
                )
            )
        );

        // Go ahead and sign off once all the transactions
        // have been inserted.
        withSignOffProposal(
            insertTxnInstructions,
            this.realmsProgramID,
            REALMS_PROGRAM_VERSION_V2,
            governance.account.realm,
            governance.pubkey,
            proposalAddress,
            this.provider.publicKey,
            signatoryRecord,
            undefined
        );

        // TODO: Could vote automatically from
        // payer wallet

        let addInstructionsTxn = new sol.Transaction();
        insertTxnInstructions.forEach((i) => addInstructionsTxn.add(i));
        returnedTxns.push(addInstructionsTxn);

        /*
        // dump txn for simulation
        addInstructionsTxn.recentBlockhash = (
            await this.provider.connection.getLatestBlockhash()
        ).blockhash;
        addInstructionsTxn.feePayer = this.provider.wallet.publicKey;
        console.log(
            (await this.provider.wallet.signTransaction(addInstructionsTxn))
                .serializeMessage()
                .toString("base64")
        );
        */
        const { blockhash } =
            await this.provider.connection.getLatestBlockhash();

        return {
            txns: returnedTxns.map((txn) => {
                txn.recentBlockhash = blockhash;
                txn.feePayer = this.provider.wallet.publicKey;
                return txn;
            }),
            url: `https://app.realms.today/dao/${governance.account.realm.toString()}/proposal/${proposalAddress.toString()}?cluster=${explorerCluster(
                this.provider.connection.rpcEndpoint
            )}`,
        };
    }
}

export async function createWorkgroup(
    provider: AnchorProvider,
    proposer: Proposer,
    groupID: string,
    groupName: string,
    depositingLicenseTokens: PublicKey
): Promise<ProposalDetails> {
    if (groupName.length <= 2)
        throw "Groupname needs to be longer than 2 characters";

    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);
    const licenseMint = getLicenseMint(provider);

    const [group] = await anchor.web3.PublicKey.findProgramAddress(
        [Buffer.from(groupID), Buffer.from("work_group")],
        program.programId
    );

    const [groupLicenseTokens] = await anchor.web3.PublicKey.findProgramAddress(
        [group.toBuffer(), Buffer.from("license_tokens")],
        program.programId
    );

    const groupAuthority = await proposer.payer();

    const createWorkGroupInstruction = await program.methods
        .createWorkGroup(groupName, groupID, "http://signal.daonetes.org:8080")
        .accounts({
            groupAuthority,
            group,
            systemProgram: SystemProgram.programId,
            licenseMint,
            depositingLicenseTokens,
            groupLicenseTokens,
        })
        .instruction();

    const details = await proposer.proposeTxn(
        [createWorkGroupInstruction],
        `Create DAOnetes workgroup id="${groupID}" name="${groupName}"`,
        `https://todo.daonetes.com`
    );
    await provider.wallet.signAllTransactions(details.txns);
    return details;
}

export async function closeDeployment(
    provider: AnchorProvider,
    proposer: Proposer,
    workGroup: PublicKey,
    deployment: PublicKey
) {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);

    const [deploymentMint] = await anchor.web3.PublicKey.findProgramAddress(
        [deployment.toBuffer(), Buffer.from("deployment_mint")],
        program.programId
    );
    const [deploymentTokens] = await anchor.web3.PublicKey.findProgramAddress(
        [deployment.toBuffer(), Buffer.from("deployment_tokens")],
        program.programId
    );
    const inst = await program.methods
        .closeDeployment()
        .accounts({
            groupAuthority: await proposer.payer(),
            deployment,
            deploymentMint,
            deploymentTokens,
            workGroup,
            systemProgram: SystemProgram.programId,
            tokenProgram: TOKEN_PROGRAM_ID,
            rent: SYSVAR_RENT_PUBKEY,
        })
        .instruction();

    const details = await proposer.proposeTxn(
        [inst],
        `Close DAOnetes deployment work_group="${workGroup.toString()}" deployment="${deployment.toString()}"`,
        `https://todo.daonetes.com`
    );
    await provider.wallet.signAllTransactions(details.txns);
    return details;
}

export async function closeSpec(
    provider: AnchorProvider,
    proposer: Proposer,
    workGroup: PublicKey,
    spec: PublicKey
) {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);
    const inst = await program.methods
        .closeWorkSpec()
        .accounts({
            groupAuthority: await proposer.payer(),
            spec,
            workGroup,
            systemProgram: SystemProgram.programId,
        })
        .instruction();

    const details = await proposer.proposeTxn(
        [inst],
        `Close DAOnetes spec work_group="${workGroup.toString()}" spec="${spec.toString()}"`,
        `https://todo.daonetes.com`
    );
    await provider.wallet.signAllTransactions(details.txns);
    return details;
}

export async function closeDevice(
    provider: AnchorProvider,
    proposer: Proposer,
    workGroup: PublicKey,
    devicePDA: PublicKey
) {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);

    const inst = await program.methods
        .closeDevice()
        .accounts({
            groupAuthority: await proposer.payer(),
            device: devicePDA,
            workGroup,
            systemProgram: SystemProgram.programId,
        })
        .instruction();

    const details = await proposer.proposeTxn(
        [inst],
        `Close device in workgroup group_id="${workGroup.toString()}" devicePDA="${devicePDA.toString()}"`,
        `https://todo.daonetes.com`
    );
    await provider.wallet.signAllTransactions(details.txns);
    return details;
}

export async function closeWorkGroup(
    provider: AnchorProvider,
    proposer: Proposer,
    workGroup: PublicKey
) {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);
    const licenseMint = getLicenseMint(provider);
    const withdrawingLicenseTokens = await getAssociatedTokenAddress(
        licenseMint,
        await proposer.payer(),
        true
    );

    const [groupLicenseTokens] = await anchor.web3.PublicKey.findProgramAddress(
        [workGroup.toBuffer(), Buffer.from("license_tokens")],
        program.programId
    );

    const groupAuthority = await proposer.payer();

    const inst = await program.methods
        .closeWorkGroup(true)
        .accounts({
            groupAuthority,
            group: workGroup,
            licenseMint,
            groupLicenseTokens,
            withdrawingLicenseTokens,
            systemProgram: SystemProgram.programId,
            tokenProgram: TOKEN_PROGRAM_ID,
            rent: SYSVAR_RENT_PUBKEY,
        })
        .instruction();

    const details = await proposer.proposeTxn(
        [inst],
        `Close DAOnetes workgroup id="${workGroup.toString()}"`,
        `https://todo.daonetes.com`
    );
    await provider.wallet.signAllTransactions(details.txns);
    return details;
}

async function sha256Digest(text: string): Promise<string> {
    const msgUint8 = new TextEncoder().encode(text);
    const hashBuffer = await crypto.subtle.digest("SHA-256", msgUint8);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray
        .map((b) => b.toString(16).padStart(2, "0"))
        .join("");
    return hashHex;
}

export async function createSpec(
    provider: AnchorProvider,
    proposer: Proposer,
    workGroupKey: PublicKey,
    specName: string,
    specURL: string,
    metadataURL: string
) {
    // get sha256(YAML contents), this is
    // stored in spec on chain for integrity purposes
    const getSpecResponse = await fetch(specURL);
    const body = await getSpecResponse.text();
    const specDigest = await sha256Digest(body);

    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);

    // eslint-disable-next-line no-unused-vars
    const [spec, _specPDABump] = await anchor.web3.PublicKey.findProgramAddress(
        [workGroupKey.toBuffer(), Buffer.from(specName), Buffer.from("spec")],
        program.programId
    );

    const createSpecInstruction = await program.methods
        .createWorkSpec(
            specName,
            { dockerCompose: {} },
            specURL,
            specDigest,
            metadataURL,
            false
        )
        .accounts({
            groupAuthority: await proposer.payer(),
            spec,
            workGroup: workGroupKey,
            systemProgram: SystemProgram.programId,
        })
        .instruction();

    const details = await proposer.proposeTxn(
        [createSpecInstruction],
        `Create spec in workgroup group_id="${workGroupKey.toString()}" specURL="${specURL}"`,
        `https://todo.daonetes.com`
    );
    await provider.wallet.signAllTransactions(details.txns);
    return details;
}

export async function createDeployment(
    provider: AnchorProvider,
    proposer: Proposer,
    workgroupKey: PublicKey,
    specName: string,
    deploymentName: string,
    replicas: number
) {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);

    // eslint-disable-next-line no-unused-vars
    const [spec, _specBump] = await anchor.web3.PublicKey.findProgramAddress(
        [
            workgroupKey.toBuffer(),
            Buffer.from(specName), // TODO: this worries me, as there's nothing enforced unique about the spec.name
            Buffer.from("spec"),
        ],
        program.programId
    );

    // eslint-disable-next-line no-unused-vars
    const [deployment, _deploymentBump] =
        await anchor.web3.PublicKey.findProgramAddress(
            [
                workgroupKey.toBuffer(),
                Buffer.from(deploymentName), // TODO: this worries me, as there's nothing enforced unique about the spec.name
                Buffer.from("deployment"),
            ],
            program.programId
        );

    // eslint-disable-next-line no-unused-vars
    const [deploymentMint, _deploymentMintBump] =
        await anchor.web3.PublicKey.findProgramAddress(
            [deployment.toBuffer(), Buffer.from("deployment_mint")],
            program.programId
        );

    // eslint-disable-next-line no-unused-vars
    const [deploymentTokens, _deploymentTokensBump] =
        await anchor.web3.PublicKey.findProgramAddress(
            [deployment.toBuffer(), Buffer.from("deployment_tokens")],
            program.programId
        );

    const createDeploymentInstruction = await program.methods
        .createDeployment(deploymentName, replicas)
        .accounts({
            groupAuthority: await proposer.payer(),
            workGroup: workgroupKey,
            deployment,
            spec,
            deploymentMint,
            deploymentTokens,
            systemProgram: SystemProgram.programId,
            tokenProgram: TOKEN_PROGRAM_ID,
            rent: SYSVAR_RENT_PUBKEY,
        })
        .instruction();

    const details = await proposer.proposeTxn(
        [createDeploymentInstruction],
        `Create DAOnetes deployment group_id="${workgroupKey.toString()}" from_spec="${specName}"`,
        `https://todo.daonetes.com`
    );
    await provider.wallet.signAllTransactions(details.txns);
    return details;
}

export async function registerDevice(
    provider: AnchorProvider,
    proposer: Proposer,
    groupAuthority: PublicKey,
    groupID: string,
    deviceKey: string
) {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);

    // TODO: test if the payer can afford the operation.
    // TODO: Validate that the device key is base58 and on curve and not,
    //       e.g., a container ID or PDA
    const devicePublicKey = new PublicKey(deviceKey);

    // eslint-disable-next-line no-unused-vars
    const [workGroup, _workGroupBump] =
        await anchor.web3.PublicKey.findProgramAddress(
            [Buffer.from(groupID), Buffer.from("work_group")],
            program.programId
        );

    const licenseMint = getLicenseMint(provider);

    const [groupLicenseTokens] = await anchor.web3.PublicKey.findProgramAddress(
        [workGroup.toBuffer(), Buffer.from("license_tokens")],
        program.programId
    );

    // eslint-disable-next-line no-unused-vars
    const [devicePDA, _devicePDABump] =
        await anchor.web3.PublicKey.findProgramAddress(
            [devicePublicKey.toBuffer()],
            program.programId
        );

    const registerDeviceInst = await program.methods
        .registerDevice(devicePublicKey)
        .accounts({
            groupAuthority,
            device: devicePDA,
            workGroup,
            licenseMint,
            groupLicenseTokens,
            systemProgram: SystemProgram.programId,
        })
        .instruction();

    const details = await proposer.proposeTxn(
        [
            sol.SystemProgram.transfer({
                fromPubkey: await proposer.payer(),
                toPubkey: devicePublicKey,
                lamports: LAMPORTS_PER_SOL * 0.1,
            }),
            registerDeviceInst,
            // TODO: Should get rent exempt minimum for empty account
            // and send that + txn fee...
            // but really some extra SOL should probably be sent anyway
            // so the device can keep updating metadata like IP if needed],
        ],
        `Register device in workgroup group_id="${groupID}" device_pubkey="${deviceKey}"`,
        `https://todo.daonetes.com`
    );
    await provider.wallet.signAllTransactions(details.txns);
    return details;
}

export async function scheduleDeployment(
    provider: AnchorProvider,
    proposer: Proposer,
    workGroupKey: PublicKey,
    deploymentName: string,
    deviceAuthority: PublicKey
) {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);

    // eslint-disable-next-line no-unused-vars
    const [deployment, _deploymentBump] =
        await anchor.web3.PublicKey.findProgramAddress(
            [
                workGroupKey.toBuffer(),
                Buffer.from(deploymentName),
                Buffer.from("deployment"),
            ],
            program.programId
        );

    // eslint-disable-next-line no-unused-vars
    const [deploymentMint, _deploymentMintBump] =
        await anchor.web3.PublicKey.findProgramAddress(
            [deployment.toBuffer(), Buffer.from("deployment_mint")],
            program.programId
        );

    // eslint-disable-next-line no-unused-vars
    const [deploymentTokens, _deploymentTokensBump] =
        await anchor.web3.PublicKey.findProgramAddress(
            [deployment.toBuffer(), Buffer.from("deployment_tokens")],
            program.programId
        );

    // eslint-disable-next-line no-unused-vars
    const [deviceTokens, _deviceTokensBump] =
        await anchor.web3.PublicKey.findProgramAddress(
            [
                deviceAuthority.toBuffer(),
                deployment.toBuffer(),
                Buffer.from("device_tokens"),
            ],
            program.programId
        );

    // eslint-disable-next-line no-unused-vars
    const [device, _deviceBump] =
        await anchor.web3.PublicKey.findProgramAddress(
            [deviceAuthority.toBuffer()],
            program.programId
        );

    const scheduleInstruction = await program.methods
        .schedule(1)
        .accounts({
            groupAuthority: await proposer.payer(),
            deployment,
            deploymentMint,
            deploymentTokens,
            device,
            deviceAuthority,
            deviceTokens,
            workGroup: workGroupKey,
            systemProgram: SystemProgram.programId,
            tokenProgram: TOKEN_PROGRAM_ID,
            rent: SYSVAR_RENT_PUBKEY,
        })
        .instruction();

    const details = await proposer.proposeTxn(
        [scheduleInstruction],
        `Schedule DAOnetes deployment token group_id="${workGroupKey.toString()}" token_mint="${deploymentMint.toString()}"`,
        `https://todo.daonetes.com`
    );
    await provider.wallet.signAllTransactions(details.txns);
    return details;
}
