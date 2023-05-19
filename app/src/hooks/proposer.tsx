import React from "react";
import { useContext } from "react";
import { Proposer } from "../worknet/createWorkgroup";

interface proposerContextState {
    proposer: Proposer | undefined;
    // eslint-disable-next-line no-unused-vars
    setProposer: (p: Proposer | undefined) => void;
}

export const ProposerContext = React.createContext<proposerContextState>({
    proposer: undefined,
    // eslint-disable-next-line no-unused-vars
    setProposer: (p: Proposer | undefined) => {},
});

export function useProposer() {
    return useContext(ProposerContext);
}
