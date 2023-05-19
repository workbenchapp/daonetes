import type { NextPage } from "next";

import { OnBoardingWizard } from "../components/OnBoardingWizard";
import { useWorkgroup } from "../hooks/selectedWorkgroup";
import { SpecLibraryList } from "../components/SpecLibraryList";

const Specs: NextPage = () => {
    const workgroupKey = useWorkgroup();

    if (!workgroupKey) {
        return (
            <OnBoardingWizard />
        )
    }


    if (!workgroupKey) {
        return (
            <div>new user onboarding wizard. connect wallet, select group owner, go. plus info on what it does.</div>
        )
    }
    return (
        <div>
            <h2>Spec Library</h2>
            <SpecLibraryList />
        </div>
    )
};

export default Specs;
