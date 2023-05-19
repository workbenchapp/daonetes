// IDK why Anchor isn't generating something like this...

import { PublicKey } from "@solana/web3.js";
import BN from "bn.js";

export type WorkGroup = {
    //workgroup: {"bump":255,"groupAuthority":"9tVLKVDWFD6CnZdat2QM1v9ev4673u28GDrbNVsMRGYi","specs":["13VSwbDuTbxMrbtKvmxmb9yqbt9c2SJTJrnCSVxZJNdd"],"devices":["9UkVvisYDdMGyh3Bt4gRZmP2VPv9xZ5ZrpDH464xf5H2"],"deployments":["GECxA6i75AZJzu1w1ax9Xx3D6eyaPvHM9khPuU9wh8LX"],"name":"default"}
    bump: number;
    groupAuthority: PublicKey;
    specs: PublicKey[];
    devices: PublicKey[];
    deployments: PublicKey[];
    name: string;
    identifier: string;
    chain?: { PubKey: PublicKey };
};

// SPEC: { "name": "nginx", "containers": [{ "image": "nginx", "name": "web", "args": [], "portMappings": [{ "innerPort": 80, "outerPort": 8080, "portType": { "tcp": {} } }], "mounts": [{ "mountType": { "bind": {} }, "src": "./templates", "dest": "/etc/nginx/templates" }] }], "volumes": []; }
// Stringly subject to change RSN :)
export interface SpecType {
    name: string;
    createdAt: BN;
    modifiedAt: BN;
    workType: any;
    urlOrContents: string;
    contentsSha256: string;
    metadataUrl: string;
    mutable: boolean;
}

// DEPLOYMENT: { "spec": "13VSwbDuTbxMrbtKvmxmb9yqbt9c2SJTJrnCSVxZJNdd", "name": "nginx_test_run", "args": [], "replicas": 1, "selfBump": 0, "mintBump": 255, "tokensBump": 255, "deploymentBump": 252; }
export interface DeploymentType {
    spec: PublicKey;
    name: string;
    args: [];
    replicas: number;
    chain?: { PubKey: PublicKey };
}
// const DEVICE = {
//     "ipv4": [59, 167, 214, 59],
//     "hostname": "p1",
//     "bump": 253,
//     "status": { "registered": {} },
//     "deviceAuthority": "2YuBYJ93yJXZkoxTWXJZkFR5SfRXvPiBGWc48CFePxgw",
//     "groupAuthority": "9tVLKVDWFD6CnZdat2QM1v9ev4673u28GDrbNVsMRGYi"
// };
export interface DeviceType {
    hostname: string;
    ipv4: number[];
    status: any;
    deviceAuthority: PublicKey;
    groupAuthority: PublicKey;
    chain?: { PubKey: PublicKey };
}

// const parsedMintInfo = {
//     "accountId": "5W29MqWG9TgjAsV3XkKAhEDdrENK95KXCFFqZ9zJ6pjh",
//     "accountInfo": {
//         "data": {
//             "parsed": {
//                 "info": {
//                     "decimals": 0,
//                     "freezeAuthority": null,
//                     "isInitialized": true,
//                     "mintAuthority": "DNWPrcGyswAv82YkRfvvj2NTP5Lbveinmu5dxUM6Jyow",
//                     "supply": "3"
//                 },
//                 "type": "mint"
//             },
//             "program": "spl-token",
//             "space": 82
//         },
//         "executable": false,
//         "lamports": 1461600,
//         "owner": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
//         "rentEpoch": 0
//     },
//     "pubKey": "5W29MqWG9TgjAsV3XkKAhEDdrENK95KXCFFqZ9zJ6pjh",
//     "net": "10.10.10.190",
//     "count": 0,
//     "solDelta": 0,
//     "maxDelta": 0
// }
export interface MintDataType {
    data: {
        parsed: {
            type: string;
            info: {
                decimals: number;
                freezeAuthority: PublicKey;
                isInitialized: boolean;
                mintAuthority: PublicKey;
                supply: number;
            };
        };
        program: string;
        space: number;
    };
    executable: boolean;
    lamports: number;
    owner: PublicKey;
    rentEpoch: number;
}
