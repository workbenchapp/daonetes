import type { NextPage } from "next";

import { WorkGroupStatus } from "../components/WorkGroupStatus";
import { OnBoardingWizard } from "../components/OnBoardingWizard";
import React from "react";


import { useWorkgroup } from "../hooks/selectedWorkgroup";


const Workgroup: NextPage = () => {
    const workgroupKey = useWorkgroup();

    if (!workgroupKey) {
        return (
            <OnBoardingWizard />
        )
    }

    return (
        <WorkGroupStatus workgroupKey={workgroupKey} />
    );
};

export default Workgroup;
