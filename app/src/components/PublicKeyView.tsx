
import { PublicKey } from '@solana/web3.js';
import { useState } from 'react';
import { OverlayTrigger, Tooltip } from 'react-bootstrap';

// https://react-icons.github.io/react-icons/icons?name=ai
import { AiOutlineCopy } from 'react-icons/ai';

// Copied from solana-workbench
function CopyIcon(props: { writeValue: string; }) {
    const { writeValue } = props;
    const [copyTooltipText, setCopyTooltipText] = useState<string | undefined>(
        'Copy'
    );

    const renderCopyTooltip = (id: string) =>
        // eslint-disable-next-line react/display-name
        function (ttProps: any) {
            return (
                <Tooltip id={id} {...ttProps}>
                    <div>{copyTooltipText}</div>
                </Tooltip>
            );
        };

    return (
        <OverlayTrigger
            placement="bottom"
            delay={{ show: 250, hide: 0 }}
            overlay={renderCopyTooltip('rootKey')}
        >
            <span
                onClick={(e) => {
                    e.stopPropagation();
                    setCopyTooltipText('Copied!');
                    // WOW: this is almost supported everywhere - https://blog.logrocket.com/implementing-copy-clipboard-react-clipboard-api/
                    navigator.clipboard.writeText(writeValue);
                }}
                onMouseLeave={(

                ) => window.setTimeout(() => setCopyTooltipText('Copy'), 500)}
                className="icon-interactive p-2 hover:bg-contrast/10 rounded-full inline-flex items-center justify-center cursor-pointer"
            >
                <AiOutlineCopy />
            </span>
        </OverlayTrigger>
    );
}

const prettifyPubkey = (pkObj: string | any = '', formatLength?: number) => {
    const pk = pkObj.toString();
    if (pk === null) {
        // cope with bad data in config
        return '';
    }
    if (pk === "none") {
        // cope with bad data in config
        return '';
    }
    if (!formatLength || formatLength + 2 > pk.length) {
        return pk;
    }
    const partLen = (formatLength - 1) / 2;

    return `${pk.slice(0, partLen)}…${pk.slice(pk.length - partLen, pk.length)}`;
};

interface PublicKeyViewProps {
    publicKey: PublicKey | undefined
}
// can be anything that has a toString...
export function PublicKeyView({ publicKey }: PublicKeyViewProps) {
    const pretty = prettifyPubkey(publicKey, 10);
    return (
        <>
            <span className="code">{pretty}</span>
            <CopyIcon writeValue={publicKey?.toString() || ""} />
        </>
    );
}