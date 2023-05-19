import { Program } from "@project-serum/anchor";

const { TOKEN_PROGRAM_ID } = require("@solana/spl-token");
const anchor = require("@project-serum/anchor");
const { assert } = require("chai");
const { SystemProgram, SYSVAR_RENT_PUBKEY } = anchor.web3;
import { Worknet } from "../target/types/worknet";

describe("app-basics", () => {
    const provider = anchor.AnchorProvider.env();
    anchor.setProvider(provider);
    const program = anchor.workspace.Worknet as Program<Worknet>;
    const payer = anchor.web3.Keypair.fromSecretKey(
        new Uint8Array(JSON.parse(process.env.WORKNET_TEST_PAYER))
    );

    // seed has max limit
    const groupIdentifier = payer.publicKey.toString().slice(0, 16);

    const specName = "test";
    const villain = anchor.web3.Keypair.generate();
    const deviceKey = anchor.web3.Keypair.generate();
    const licenseMint = new anchor.web3.PublicKey(
        "3CrKoTYzfbeenzQmhXMQzHM929kioTvsr1JtDSX9uET5"
    );
    const depositingLicenseTokens = new anchor.web3.PublicKey(
        "9RNxHPFffWYcnAf2BN521VzUmCFhXNGoMKm1BdrZ3JqE"
    );
    const deploymentName = "nginx_with_sidecar";

    // anchor.utils.features.set("debug-logs");

    it("Initialize program state", async () => {
        await provider.connection.confirmTransaction(
            await provider.connection.requestAirdrop(
                payer.publicKey,
                10000000000
            ),
            "processed"
        );
        await provider.connection.confirmTransaction(
            await provider.connection.requestAirdrop(
                villain.publicKey,
                10000000000
            ),
            "processed"
        );
    });

    it("creates a work group", async () => {
        const [group, groupBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [Buffer.from(groupIdentifier), Buffer.from("work_group")],
                program.programId
            );

        const [groupLicenseTokens, _groupLicenseTokensBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [group.toBuffer(), Buffer.from("license_tokens")],
                program.programId
            );

        const customSignalURL = "https://customsignalserver.com";
        await program.methods
            .createWorkGroup("default", groupIdentifier, customSignalURL)
            .accounts({
                groupAuthority: provider.wallet.publicKey,
                group,
                licenseMint,
                depositingLicenseTokens,
                groupLicenseTokens,
                systemProgram: SystemProgram.programId,
            })
            .rpc();

        const groupAccount = await program.account.workGroup.fetch(group);
        assert.equal(groupAccount.specs.length, 0);
        assert.equal(groupAccount.bump, groupBump);
        assert.equal(
            groupAccount.groupAuthority.toString(),
            provider.wallet.publicKey.toString()
        );
        assert.equal(groupAccount.signalServerUrl.toString(), customSignalURL);
    });

    it("creates a work spec", async () => {
        const [workGroup, _workGroupBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [Buffer.from(groupIdentifier), Buffer.from("work_group")],
                program.programId
            );
        const [spec, _specPDABump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [
                    workGroup.toBuffer(),
                    Buffer.from(specName),
                    Buffer.from("spec"),
                ],
                program.programId
            );

        await program.methods
            .createWorkSpec(
                specName,
                { dockerCompose: {} },
                "https://raw.githubusercontent.com/docker/awesome-compose/master/elasticsearch-logstash-kibana/compose.yaml",
                "09409df7d8b9cac06f9ebdc2bcdf11155d924307f2354f9de2973b81007bacc0",
                "",
                false
            )
            .accounts({
                groupAuthority: provider.wallet.publicKey,
                spec,
                workGroup,
                systemProgram: SystemProgram.programId,
            })
            .rpc();

        const fetchedSpec = await program.account.workSpec.fetch(spec);
        assert.deepEqual(specName, fetchedSpec.name);
        // TODO: verify other spec properties
        const fetchedGroup = await program.account.workGroup.fetch(workGroup);
        assert.equal(fetchedGroup.specs.length, 1);
        assert.equal(fetchedGroup.specs[0].toString(), spec.toString());
    });

    it("creates a deployment", async () => {
        const deploymentName = "nginx_with_sidecar";

        const [workGroup, _workGroupBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [Buffer.from(groupIdentifier), Buffer.from("work_group")],
                program.programId
            );

        const [spec, _specBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [
                    workGroup.toBuffer(),
                    Buffer.from(specName),
                    Buffer.from("spec"),
                ],
                program.programId
            );

        const [deployment, _deploymentBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [
                    workGroup.toBuffer(),
                    Buffer.from(deploymentName),
                    Buffer.from("deployment"),
                ],
                program.programId
            );

        const [deploymentMint, _deploymentMintBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [deployment.toBuffer(), Buffer.from("deployment_mint")],
                program.programId
            );

        const [deploymentTokens, _deploymentTokensBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [deployment.toBuffer(), Buffer.from("deployment_tokens")],
                program.programId
            );

        await program.methods
            .createDeployment(deploymentName, 3)
            .accounts({
                groupAuthority: provider.wallet.publicKey,
                workGroup,
                deployment,
                spec,
                deploymentMint,
                deploymentTokens,
                systemProgram: SystemProgram.programId,
                tokenProgram: TOKEN_PROGRAM_ID,
                rent: SYSVAR_RENT_PUBKEY,
            })
            .rpc();
    });

    // TODO
    it("shouldn't allow work requests to be created from keys not matching the workgroup", async () => {});

    it("registers a device", async () => {
        const [devicePDA, _devicePDABump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [deviceKey.publicKey.toBuffer()],
                program.programId
            );

        const [workGroup, _workGroupBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [Buffer.from(groupIdentifier), Buffer.from("work_group")],
                program.programId
            );
        const [groupLicenseTokens, _groupLicenseTokensBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [workGroup.toBuffer(), Buffer.from("license_tokens")],
                program.programId
            );

        await program.methods
            .registerDevice(deviceKey.publicKey)
            .accounts({
                groupAuthority: provider.wallet.PublicKey,
                device: devicePDA,
                workGroup,
                licenseMint,
                groupLicenseTokens,
                systemProgram: SystemProgram.programId,
                tokenProgram: TOKEN_PROGRAM_ID,
            })
            .rpc();
        const fetchedDevice = await program.account.device.fetch(devicePDA);
        assert.equal(
            deviceKey.publicKey.toString(),
            fetchedDevice.deviceAuthority.toString()
        );
        assert.equal(fetchedDevice.workGroup.toString(), workGroup.toString());
        const fetchedGroup = await program.account.workGroup.fetch(workGroup);
        assert.equal(fetchedGroup.devices.length, 1);
    });

    it("updates device info", async () => {
        const [devicePDA, devicePDABump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [deviceKey.publicKey.toBuffer()],
                program.programId
            );
        await program.methods
            .updateDevice([10, 10, 10, 10], "testhost", devicePDABump, {
                registered: {},
            })
            .accounts({
                deviceAuthority: deviceKey.publicKey,
                device: devicePDA,
            })
            .signers([deviceKey])
            .rpc();
        const fetchedDevice = await program.account.device.fetch(devicePDA);
        assert.deepEqual([10, 10, 10, 10], fetchedDevice.ipv4);
        assert.equal("testhost", fetchedDevice.hostname);
        assert.equal(devicePDABump, fetchedDevice.bump);
    });

    it("schedules a deployment", async () => {
        const [workGroup, _workGroupBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [Buffer.from(groupIdentifier), Buffer.from("work_group")],
                program.programId
            );

        const [deployment, _deploymentBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [
                    workGroup.toBuffer(),
                    Buffer.from(deploymentName),
                    Buffer.from("deployment"),
                ],
                program.programId
            );

        const [deploymentMint, _deploymentMintBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [deployment.toBuffer(), Buffer.from("deployment_mint")],
                program.programId
            );

        const [deploymentTokens, _deploymentTokensBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [deployment.toBuffer(), Buffer.from("deployment_tokens")],
                program.programId
            );

        const [device, _devicePDABump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [deviceKey.publicKey.toBuffer()],
                program.programId
            );

        const [deviceTokens, _deviceTokensBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [
                    deviceKey.publicKey.toBuffer(),
                    deployment.toBuffer(),
                    Buffer.from("device_tokens"),
                ],
                program.programId
            );

        await program.methods
            .schedule(1)
            .accounts({
                groupAuthority: provider.wallet.publicKey,
                deployment,
                deploymentMint,
                deploymentTokens,
                device,
                deviceAuthority: deviceKey.publicKey.toString(),
                deviceTokens,
                workGroup,
                systemProgram: SystemProgram.programId,
                tokenProgram: TOKEN_PROGRAM_ID,
                rent: SYSVAR_RENT_PUBKEY,
            })
            .rpc();

        // todo: check that token balance in device's token
        // account is now 3
    });

    it("closes a deployment", async () => {
        const [workGroup, _workGroupBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [Buffer.from(groupIdentifier), Buffer.from("work_group")],
                program.programId
            );

        const [deployment, _deploymentBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [
                    workGroup.toBuffer(),
                    Buffer.from(deploymentName),
                    Buffer.from("deployment"),
                ],
                program.programId
            );

        const [deploymentMint, _deploymentMintBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [deployment.toBuffer(), Buffer.from("deployment_mint")],
                program.programId
            );

        const [deploymentTokens, _deploymentTokensBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [deployment.toBuffer(), Buffer.from("deployment_tokens")],
                program.programId
            );

        await program.methods
            .closeDeployment()
            .accounts({
                groupAuthority: provider.wallet.publicKey,
                deployment,
                deploymentMint,
                deploymentTokens,
                workGroup,
                systemProgram: SystemProgram.programId,
                tokenProgram: TOKEN_PROGRAM_ID,
                rent: SYSVAR_RENT_PUBKEY,
            })
            .rpc();
        const fetchedGroup = await program.account.workGroup.fetch(workGroup);
        assert.equal(fetchedGroup.deployments.length, 0);
    });

    it("closes a spec", async () => {
        const [workGroup, _workGroupBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [Buffer.from(groupIdentifier), Buffer.from("work_group")],
                program.programId
            );

        const [spec, _specBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [
                    workGroup.toBuffer(),
                    Buffer.from(specName),
                    Buffer.from("spec"),
                ],
                program.programId
            );

        await program.methods
            .closeWorkSpec()
            .accounts({
                groupAuthority: provider.wallet.publicKey,
                spec,
                workGroup,
                systemProgram: SystemProgram.programId,
            })
            .rpc();
        const fetchedGroup = await program.account.workGroup.fetch(workGroup);
        assert.equal(fetchedGroup.specs.length, 0);
        // TODO: these close methods should really be checking refunds too,
        // but that's more work than I'd like to put in atm :P
    });

    it("closes a device", async () => {
        const [workGroup, _workGroupBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [Buffer.from(groupIdentifier), Buffer.from("work_group")],
                program.programId
            );

        const [device, _devicePDABump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [deviceKey.publicKey.toBuffer()],
                program.programId
            );

        await program.methods
            .closeDevice()
            .accounts({
                groupAuthority: provider.wallet.publicKey,
                device,
                workGroup,
                systemProgram: SystemProgram.programId,
            })
            .rpc();
        const fetchedGroup = await program.account.workGroup.fetch(workGroup);
        assert.equal(fetchedGroup.devices.length, 1);
        assert.equal(
            fetchedGroup.devices[0].equals(SystemProgram.programId),
            true
        );
    });

    it("closes a workgroup", async () => {
        const [group, _workGroupBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [Buffer.from(groupIdentifier), Buffer.from("work_group")],
                program.programId
            );

        const [groupLicenseTokens, _groupLicenseTokensBump] =
            await anchor.web3.PublicKey.findProgramAddress(
                [group.toBuffer(), Buffer.from("license_tokens")],
                program.programId
            );

        await program.methods
            .closeWorkGroup(false)
            .accounts({
                groupAuthority: provider.wallet.publicKey,
                group,
                licenseMint,
                groupLicenseTokens,
                withdrawingLicenseTokens: depositingLicenseTokens,
                systemProgram: SystemProgram.programId,
                tokenProgram: TOKEN_PROGRAM_ID,
                rent: SYSVAR_RENT_PUBKEY,
            })
            .rpc();

        try {
            await program.account.workGroup.fetch(group);
        } catch (e) {
            assert.equal(e.message.includes("Account does not exist"), true);
        }
    });
});
