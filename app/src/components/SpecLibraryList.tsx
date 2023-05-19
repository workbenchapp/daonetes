import { Button, Card, Col, Container, Image, Row } from "react-bootstrap";
import { useQuery } from "@tanstack/react-query";
import { createSpec } from "../worknet/createWorkgroup";
import { getProvider } from "../worknet/worknet";
import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { useWorkgroup } from "../hooks/selectedWorkgroup";
import { useProposer } from "../hooks/proposer";
import { toastAndSendProposedTransactions } from "../worknet/displays";

// {
//     "type": "docker-compose",
//     "name": "Caddy",
//     "description": "reverse proxy",
//     "logo": "https://caddyserver.com/resources/images/caddy-logo.svg",
//     "dependencies": [],
//     "url": "https://caddyserver.com",
//     "spec": "./caddy/docker-compose.yml",
//     "environment": [
//         {
//             "name": "DOMAIN",
//             "label": "The Top level domain to add container proxying to",
//             "description": "top level DNS that containers will get added to either using *.DOMAIN, or DOMAIN/path, or a mixture.",
//             "default": "loc.alho.st"
//         }
//     ]
// }
type Entry = {
    type: string;
    name: string;
    description: string;
    logo: string;
    dependencies: string[];
    url: string;
    spec: string;
    environment: {
        name: string;
        label: string;
        description: string;
        default: string;
    }[];
};

type Library = {
    version: string;
    specs: Entry[];
};

interface SpecLibraryItemProps {
    libraryUrl: string;
    spec: Entry;
}
function SpecLibraryItem({ libraryUrl, spec }: SpecLibraryItemProps) {
    const wallet = useWallet();
    const { connection } = useConnection();
    const workGroup = useWorkgroup();

    const provider = getProvider(connection, wallet);
    const { proposer } = useProposer();

    var specUrl = spec.spec;
    if (specUrl.startsWith("./")) {
        specUrl = libraryUrl.replace("specs.json", spec.spec.replace("./", ""));
    }

    return (
        <Col>
            <Card key={spec.name}>
                <Card.Header>
                    <Container>
                        <Row fluid>
                            <Col md={true}>{spec.name}</Col>
                            <Col className="nav ml-auto justify-content-end">
                                <Image
                                    alt={spec.name}
                                    src={spec.logo}
                                    height="32px"
                                />
                            </Col>
                        </Row>
                    </Container>
                </Card.Header>
                <Card.Body>
                    <Card.Text>
                        <span>{spec.description}</span>
                        <br />
                        <span>
                            <a href={spec.url}>Project URL</a>
                        </span>
                    </Card.Text>
                </Card.Body>
                <Card.Footer>
                    {/* TODO: disable if already imported..
                TODO: or switch to deploy.... */}
                    <Button
                        disabled={!workGroup ? true : false}
                        onClick={async () => {
                            if (!workGroup || !proposer) return;
                            toastAndSendProposedTransactions(
                                "CreateWorkSpec",
                                proposer,
                                provider.connection,
                                await createSpec(
                                    provider,
                                    proposer,
                                    workGroup,
                                    spec.name,
                                    spec.url,
                                    "" // Metadata URL
                                )
                            );
                        }}
                    >
                        {proposer?.dao() && "Propose "}Import Spec
                    </Button>
                    <span>
                        <a href={specUrl}>Compose File</a>
                    </span>
                </Card.Footer>
            </Card>
        </Col>
    );
}

export function SpecLibraryList() {
    // TODO: will be building out a Dapp that serves this from a URL
    const specUrl =
        "https://raw.githubusercontent.com/workbenchapp/library/main/specs.json";

    const {
        status,
        data: specs,
        error,
    } = useQuery<Library>(["specs"], async () => {
        const response = await fetch(specUrl);

        if (!response.ok) {
            throw new Error("Network response was not ok");
        }

        return response.json();
    });

    if (!specs) {
        var errorMsg = "";
        if (error && error instanceof Error) {
            errorMsg = error.message; // ok
        } else {
            errorMsg = JSON.stringify(error);
        }
        return (
            <div>
                query: specs fetch {status}: {errorMsg}
            </div>
        );
    }

    return (
        <Row xs={1} md={2} className="g-4">
            {specs.specs.map((spec) => {
                return (
                    <SpecLibraryItem
                        libraryUrl={specUrl}
                        spec={spec}
                        key={spec.name}
                    />
                );
            })}
        </Row>
    );
}
