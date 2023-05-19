import Head from "next/head";

import React, { useMemo, useState, useEffect } from "react";

import Container from "react-bootstrap/Container";
import Navbar from "react-bootstrap/Navbar";
import NavDropdown from "react-bootstrap/NavDropdown";

import {
    ConnectionProvider,
    useConnection,
    useWallet,
    WalletProvider,
} from "@solana/wallet-adapter-react";
import {
    CoinbaseWalletAdapter,
    GlowWalletAdapter,
    PhantomWalletAdapter,
    SlopeWalletAdapter,
    SolflareWalletAdapter,
    // SolletExtensionWalletAdapter,
    // SolletWalletAdapter,
    TorusWalletAdapter,
} from "@solana/wallet-adapter-wallets";
import {
    WalletModalProvider,
    WalletMultiButton,
} from "@solana/wallet-adapter-react-ui";
import { LAMPORTS_PER_SOL } from "@solana/web3.js";
// import {
//     createDefaultAddressSelector,
//     createDefaultAuthorizationResultCache,
//     SolanaMobileWalletAdapter,
// } from "@solana-mobile/wallet-adapter-mobile";
import { useQuery } from "@tanstack/react-query";

import packageJson from "../../package.json";

//import { testGokiWallet } from "../worknet/worknet";
import {
    useValidatorNetwork,
    useValidatorNetworkName,
} from "../hooks/validatorNetwork";
import { DaonetesWorkGroup } from "./WorkgroupsList";
import { useWorkgroup } from "../hooks/selectedWorkgroup";
import { Nav } from "react-bootstrap";
import { ToastContainer } from "react-toastify";

import Image from "next/image";

import LogoImage from "../../public/icon.jpg";

import { ProposerContext } from "../hooks/proposer";
import { LocalWalletProposer, Proposer } from "../worknet/createWorkgroup";
import { getProvider } from "../worknet/worknet";

// Default styles that can be overridden by your app
require("@solana/wallet-adapter-react-ui/styles.css");

import { loadIntercom, initIntercomWindow, updateIntercom } from "next-intercom";

loadIntercom({
    appId: "ovpbqqsq", // default : ''
    //email: "", //default: ''
    //name: "", //default: RandomName
    ssr: false, // default: false
    initWindow: true, // default: true
    delay: 0, // default: 0  - usefull for mobile devices to prevent blocking the main thread
});

// If init was set to false initiate the window when needed
initIntercomWindow({ appId: "ovpbqqsq" });

interface ValidatorNetworkItemProps {
    name: string;
    url: string;
}
function ValidatorNetworkItem({ name, url }: ValidatorNetworkItemProps) {
    const workgroupOwner = useWorkgroup();

    const networkUrl = new URLSearchParams();

    networkUrl.set("network", url);
    if (workgroupOwner) {
        networkUrl.set("workgroup", workgroupOwner?.toString());
    }

    console.log(networkUrl.toString());

    return (
        <Nav.Link>
            <NavDropdown.Item href={"?" + networkUrl.toString()}>{name}</NavDropdown.Item>
        </Nav.Link>
    );
}

export function ValidatorNetwork() {
    const networkName = useValidatorNetworkName();

    return (
        <NavDropdown title={networkName} id="navbarScrollingDropdown" drop="up">
            <ValidatorNetworkItem name="mainnet-beta" url="mainnet-beta" />
            <ValidatorNetworkItem name="devnet" url="devnet" />
            <ValidatorNetworkItem name="testnet" url="testnet" />
            <NavDropdown.Divider />
            <ValidatorNetworkItem
                name="localnet"
                url="http://localhost:8899/"
            />
            <ValidatorNetworkItem
                name="http://10.10.10.190:8899/"
                url="http://10.10.10.190:8899/"
            />
            <NavDropdown.Divider />
        </NavDropdown>
    );
}

export function Header() {
    const workgroupOwner = useWorkgroup();
    const networkName = useValidatorNetworkName();
    const params = new URLSearchParams();

    params.set("network", networkName);
    if (workgroupOwner) {
        params.set("workgroup", workgroupOwner.toString());
    }

    if (workgroupOwner) {
        updateIntercom("update", {
            workgroup: workgroupOwner.toString(),
        });
    }

    return (
        <Navbar>
            <Container fluid>
                <Navbar.Brand href="/">
                    <Image
                        alt="The DAOnetes logo, a crystalline turquoise dragonfly"
                        src={LogoImage.src}
                        height={16}
                        width={16}
                    />
                    <span className="ms-2">{packageJson.name}</span>
                </Navbar.Brand>
                <Navbar.Toggle aria-controls="navbarScroll" />
                <Navbar.Collapse id="navbarScroll">
                    <Nav
                        className="me-auto my-2 my-lg-0"
                        style={{ maxHeight: "100px" }}
                        navbarScroll
                    >
                        <Nav.Link href={"/nodes?" + params.toString()}>Nodes</Nav.Link>
                        <Nav.Link href={"/dashboard?" + params.toString()}>Services</Nav.Link>
                        <Nav.Link href={"/workgroup?" + params.toString()}>Manage</Nav.Link>
                        <Nav.Link href={"/specs?" + params.toString()}>Library</Nav.Link>
                        <Nav.Link href={"/about?" + params.toString()}>About</Nav.Link>
                    </Nav>
                    {/* <Form className="d-flex">
                        <Form.Control
                            type="search"
                            placeholder="Search"
                            className="me-2"
                            aria-label="Search"
                        />
                        <Button variant="outline-success">Search</Button>
                    </Form> */}
                </Navbar.Collapse>
                <WalletMultiButton className="wallet-adapter-button-custom" />
            </Container>
        </Navbar>
    );
}

export function Footer() {
    const wallet = useWallet();
    const { connection } = useConnection();

    const url = connection.rpcEndpoint;
    const { status /* , error */, data } = useQuery(
        ["walletBalance", wallet.publicKey],
        async () => {
            if (!wallet.publicKey) {
                return 0;
            }
            return await connection.getBalance(wallet.publicKey);
        }
    );

    const walletSOL = status === "success" ? data / LAMPORTS_PER_SOL : 0;
    updateIntercom("update", {
        userwallet: wallet.publicKey?.toString(),
    });

    return (
        <Navbar bg="light" fixed="bottom">
            <Container>
                <Navbar.Text>
                    <div>
                        {wallet.connected ? "connected" : "offline"} to {url}
                    </div>
                    <div>Wallet has {walletSOL} SOL</div>
                </Navbar.Text>
                <DaonetesWorkGroup />
                <ValidatorNetwork />
            </Container>
        </Navbar>
    );
}

export function Sidebar() {
    return <></>;
}

type LayoutProps = {
    children: React.ReactNode;
};

function PageContainer({ children }: LayoutProps) {
    const [proposer, setProposer] = useState<Proposer | undefined>(undefined);
    const wallet = useWallet();
    const { connection } = useConnection();

    useEffect(() => {
        const provider = getProvider(connection, wallet);
        if (wallet.publicKey) {
            setProposer(new LocalWalletProposer(provider));
        }
    }, [connection, setProposer, wallet]);

    return (
        <ProposerContext.Provider value={{ proposer, setProposer }}>
            {children}
        </ProposerContext.Provider>
    );
}

export default function LayoutDefault({ children }: LayoutProps) {
    // The network can be set to 'devnet', 'testnet', or 'mainnet-beta'.
    // const network = WalletAdapterNetwork.Devnet
    // const networkUrl = clusterApiUrl(network);
    // // You can also provide a custom RPC endpoint.
    // const endpoint = useMemo(() => networkUrl, [networkUrl]);
    // const network = 'Localnet';
    // const endpoint = 'http://10.10.10.190:8899';

    const network = useValidatorNetwork();

    // @solana/wallet-adapter-wallets includes all the adapters but supports tree shaking and lazy loading --
    // Only the wallets you configure here will be compiled into your application, and only the dependencies
    // of wallets that your users connect to will be loaded.
    const wallets = useMemo(
        () => [
            // new SolanaMobileWalletAdapter({
            //     cluster: "devnet",

            //     addressSelector: createDefaultAddressSelector(),
            //     appIdentity: {
            //         name: "DAONetes",
            //         uri: "https://daonetes.org",
            //         icon: LogoImage.src,
            //     },
            //     authorizationResultCache:
            //         createDefaultAuthorizationResultCache(),
            // }),
            new CoinbaseWalletAdapter(),
            new PhantomWalletAdapter(),
            new GlowWalletAdapter(),
            new SlopeWalletAdapter(),
            new SolflareWalletAdapter({}),
            new TorusWalletAdapter(),
        ],
        [network]
    );

    return (
        <Container>
            <Head>
                <title>{packageJson.name}</title>
                <meta name="description" content={packageJson.description} />
                <link rel="icon" href="/favicon.ico" />
            </Head>
            <ConnectionProvider endpoint={network}>
                <WalletProvider wallets={wallets} autoConnect>
                    <WalletModalProvider>
                        <Header />
                        <div className="paddingFromFooter">
                            <Sidebar />
                            <PageContainer>{children}</PageContainer>
                        </div>
                        <ToastContainer position="bottom-left" />
                        <Footer />
                    </WalletModalProvider>
                </WalletProvider>
            </ConnectionProvider>
        </Container>
    );
}
