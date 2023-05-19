import { AnchorProvider } from "@project-serum/anchor";
import { TOKEN_PROGRAM_ID } from "@solana/spl-token";
import * as sol from "@solana/web3.js";
import { getNativeTreasuryAddress } from "@solana/spl-governance";
import { WorkGroup } from "./types";

// Retrieve the governance associated with a given wallet key and work group ID.
// This allows us to automatically determine if a given wallet user is a
// member of the DAO that is the authority for a selected workgroup, i.e., know
// automatically that we should use the DAO proposer.
//
// From a given workgroup ID, we know its group authority, which
// is the treasury in the case of a Realms governance. We also know
// the token accounts owned by the browser users' currently connected wallet.
//
// Since the PDA seeds for the treasury are ['treasury', governance], and the
// governance is the mint authority for its mint, we can iterate through the
// mint authorities of the token accounts to check if the PDA matches, which
// would indicate that is the correct governance.
//
// TODO: This code may be dysfunctional or broken if there are many token
// accounts owned by the wallet, or if the user somehow closes the token account
// associated with their governance tokens after depositing them into the DAO.
export async function realmsGovernanceFromWorkgroup(
    provider: AnchorProvider,
    workgroup: WorkGroup | undefined
): Promise<sol.PublicKey | undefined> {
    // TODO: There is a bug where in some circumstances, workgroup
    // is undefined even though it should not be. It's unclear if this is
    // due to development environment hot reload cleverness or something
    // else entirely.
    if (!workgroup || !provider.wallet.publicKey) return undefined;

    const tokenAccounts =
        await provider.connection.getParsedTokenAccountsByOwner(
            provider.wallet.publicKey,
            {
                programId: TOKEN_PROGRAM_ID,
            }
        );

    // hack to get access to _rpcRequest
    // (web3.js has no method for parsed multiple mints)
    const conn: any = provider.connection;
    const parsedAcctResp = await conn._rpcRequest(
        "getMultipleAccounts",
        conn._buildArgs(
            [
                tokenAccounts.value.map(
                    (tokenAccount) => tokenAccount.account.data.parsed.info.mint
                ),
            ],
            "confirmed",
            "jsonParsed",
            {}
        )
    );
    const mintAuthorities = parsedAcctResp.result.value.map(
        (acct: any) => new sol.PublicKey(acct.data.parsed.info.mintAuthority)
    );

    const governanceProgramID = new sol.PublicKey(
        "GovER5Lthms3bLBqWub97yVrMmEogzX7xNjdXpPPCVZw"
    );
    for await (const mintAuthority of mintAuthorities) {
        const treasuryPDA = await getNativeTreasuryAddress(
            // TODO: This could be a custom governance program address.
            governanceProgramID,
            mintAuthority
        );
        if (treasuryPDA.equals(workgroup.groupAuthority)) {
            return mintAuthority;
        }
    }
}
