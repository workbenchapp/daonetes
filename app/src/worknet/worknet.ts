import {
    Connection,
    PublicKey,
    TokenAccountsFilter,
    Transaction,
} from "@solana/web3.js";
import * as anchor from "@project-serum/anchor";
import { Program } from "@project-serum/anchor";
import { AnchorProvider } from "@project-serum/anchor";

import {
    ConnectionContextState,
    WalletContextState,
} from "@solana/wallet-adapter-react";

import * as idl from "../../../target/idl/worknet.json";
import { Worknet } from "../../../target/types/worknet";
import {
    DeploymentType,
    DeviceType,
    MintDataType,
    SpecType,
    WorkGroup,
} from "./types";

export const testGokiWallet = "9tVLKVDWFD6CnZdat2QM1v9ev4673u28GDrbNVsMRGYi";
export const GokiProgramPubkey = new PublicKey(
    "GokivDYuQXPZCWRkwMhdH2h91KpDQXBEmpgBgs55bnpH"
);
export const WORKNET_V1_PROGRAM_ID =
    "EdUCoDdRnT5HsQ2Ejy3TWMTQP8iUyMQB4WzoNh45pNX9";
const programID = new PublicKey(WORKNET_V1_PROGRAM_ID);

export const getWorknetProgram = (provider: AnchorProvider) => {
    // What's up with the below? It's gross, but
    // AFAICT, Anchor never really expected us to cast
    // the IDL to Worknet field directly.
    // But we wrote this prototype code this way.
    //
    // Usually program is created this way:
    //
    //     anchor.workspace.Worknet as Program<Worknet>
    //
    // So for now, to get types to be OK, (i.e., not have
    // error about instructions field not sufficiently overlapping)
    // we cast to unknown first before casting to Worknet.
    const program = new Program<Worknet>(
        idl as unknown as Worknet,
        programID,
        provider
    );
    return program;
};

const licenseMintLower = new PublicKey(
    "3CrKoTYzfbeenzQmhXMQzHM929kioTvsr1JtDSX9uET5"
);
const licenseMintUpper = new PublicKey(
    "Ew5hokTuULRDsgnhKThGv3nrw3RPjiHASQZNcRNTHJ9Z"
);

export const getLicenseMint = (connection: ConnectionContextState) => {
    if (connection.connection.rpcEndpoint === "http://localhost:8899")
        return licenseMintLower;

    // for devnet... doesn't exist on mainnet or testnet yet
    return licenseMintUpper;
};

/**
 * NOTE: tried to override the interface, but then got an error regarding Error: Signature verification failed
 * see https://github.com/coral-xyz/anchor/issues/1933
 */
type SolanaWallet = WalletContextState & {
    publicKey: PublicKey;
    // eslint-disable-next-line no-unused-vars
    signTransaction(tx: Transaction): Promise<Transaction>;
    // eslint-disable-next-line no-unused-vars
    signAllTransactions(txs: Transaction[]): Promise<Transaction[]>;
};
export function getProvider(
    connection: Connection,
    wallet: WalletContextState
): AnchorProvider {
    const provider = new AnchorProvider(connection, wallet as SolanaWallet, {
        preflightCommitment: "confirmed",
    });
    return provider;
}

export async function getWorknetGroup(
    provider: AnchorProvider,
    workgroupKey: PublicKey
): Promise<WorkGroup> {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);

    const fetchedGroup = await program.account.workGroup.fetch(workgroupKey);
    if (!fetchedGroup) {
        throw "failed to get group for " + workgroupKey;
    }
    const workGroup = fetchedGroup as WorkGroup;
    workGroup.chain = {
        PubKey: workgroupKey,
    };

    return workGroup;
}

export async function getDevice(
    provider: AnchorProvider,
    devicePDA: PublicKey
): Promise<DeviceType> {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);
    const fetchedDevice = await program.account.device.fetch(devicePDA);
    const thisDevice = fetchedDevice as DeviceType;
    thisDevice.chain = {
        PubKey: devicePDA,
    };
    console.log("DEVICE: " + JSON.stringify(thisDevice));
    return thisDevice;
}

// SPEC: { "name": "nginx", "containers": [{ "image": "nginx", "name": "web", "args": [], "portMappings": [{ "innerPort": 80, "outerPort": 8080, "portType": { "tcp": {} } }], "mounts": [{ "mountType": { "bind": {} }, "src": "./templates", "dest": "/etc/nginx/templates" }] }], "volumes": []; }
export async function getSpec(
    provider: AnchorProvider,
    specPDA: PublicKey
): Promise<SpecType> {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);
    const fetchedSpec = await program.account.workSpec.fetch(specPDA);
    const thisSpec = fetchedSpec as SpecType;

    console.log("SPEC: " + JSON.stringify(thisSpec));
    return thisSpec;
}

// DEPLOYMENT: { "spec": "13VSwbDuTbxMrbtKvmxmb9yqbt9c2SJTJrnCSVxZJNdd", "name": "nginx_test_run", "args": [], "replicas": 1, "selfBump": 0, "mintBump": 255, "tokensBump": 255, "deploymentBump": 252; }
export async function getDeployment(
    provider: AnchorProvider,
    deploymentPDA: PublicKey
): Promise<DeploymentType> {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);
    const fetchedDeployment = await program.account.deployment.fetch(
        deploymentPDA
    );
    const thisDeployment = fetchedDeployment as DeploymentType;
    thisDeployment.chain = {
        PubKey: deploymentPDA,
    };

    console.log("DEPLOYMENT: " + JSON.stringify(thisDeployment));
    return thisDeployment;
}

// should work for a deployment and for a device
export async function getAttachedTokenMints(
    provider: AnchorProvider,
    dKey: PublicKey
): Promise<MintDataType[]> {
    const filter: TokenAccountsFilter = {
        programId: new PublicKey("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
    };
    const tokenAccounts =
        await provider.connection.getParsedTokenAccountsByOwner(dKey, filter);
    const mintList: MintDataType[] = [];
    await tokenAccounts.value.map(async (tokenAccount) => {
        const mintKey = new PublicKey(
            tokenAccount.account.data.parsed.info.mint
        );
        const mint = await provider.connection.getParsedAccountInfo(mintKey);
        //console.log("minty: " + JSON.stringify(mint))
        if (mint && mint.value) {
            const mintData = mint.value as MintDataType;
            // TODO: only add it if its actually a worknet token...
            mintList.push(mintData);
        }
        return mint.value;
    });
    return mintList;
}

export async function getWorknets(provider: AnchorProvider) {
    anchor.setProvider(provider);
    const program = getWorknetProgram(provider);
    return await program.account.workGroup.all();
}
