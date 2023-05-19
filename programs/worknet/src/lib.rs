use anchor_lang::prelude::*;
use anchor_spl::token::{Mint, Token, TokenAccount};

declare_id!("EdUCoDdRnT5HsQ2Ejy3TWMTQP8iUyMQB4WzoNh45pNX9");

pub mod worknet_license_token {
    use anchor_lang::declare_id;

    #[cfg(feature = "local-license-mint")]
    declare_id!("3CrKoTYzfbeenzQmhXMQzHM929kioTvsr1JtDSX9uET5");

    #[cfg(not(feature = "local-license-mint"))]
    declare_id!("Ew5hokTuULRDsgnhKThGv3nrw3RPjiHASQZNcRNTHJ9Z");
}

#[error_code]
pub enum ErrorCode {
    #[msg("You do not have enough license tokens to perform this operation. Please deposit more")]
    InsufficentLicenseTokens,

    #[msg("Not enough replica tokens in Deployment")]
    InsuffientReplicaTokens,

    #[msg("Closing this work group would orphan specs")]
    OrphanedSpecs,

    #[msg("Closing this work group would orphan devices")]
    OrphanedDevices,

    #[msg("Closing this work group would orphan deployments")]
    OrphanedDeployments,
}

#[program]
pub mod worknet {
    use super::*;
    use anchor_lang::solana_program::{self, entrypoint::ProgramResult};
    use anchor_spl::token::{self, Approve, Burn, CloseAccount, MintTo, Transfer};

    pub fn create_work_group(
        ctx: Context<CreateWorkGroup>,
        name: String,
        identifier: String,
        signal_server_url: String,
    ) -> Result<()> {
        let depositing_license_tokens = &mut ctx.accounts.depositing_license_tokens;
        if !(depositing_license_tokens.amount > 0) {
            return err!(ErrorCode::InsufficentLicenseTokens);
        }
        let group_license_tokens = &mut ctx.accounts.group_license_tokens;

        let transfer_inst = Transfer {
            from: depositing_license_tokens.to_account_info(),
            to: group_license_tokens.to_account_info(),
            authority: ctx.accounts.group_authority.to_account_info(),
        };
        let cpi_ctx = CpiContext::new(ctx.accounts.token_program.to_account_info(), transfer_inst);
        token::transfer(cpi_ctx, 1000000000)?;

        let group = &mut ctx.accounts.group;
        group.group_authority = ctx.accounts.group_authority.key();
        group.bump = *ctx.bumps.get("group").unwrap();
        group.specs = vec![];
        group.devices = vec![];
        group.name = name;
        group.identifier = identifier;
        group.signal_server_url = signal_server_url;
        Ok(())
    }

    pub fn close_work_group(ctx: Context<CloseWorkGroup>, force: bool) -> Result<()> {
        let group = &mut ctx.accounts.group;
        if group.specs.len() > 0 && !force {
            return err!(ErrorCode::OrphanedSpecs);
        }

        for device in group.devices.iter() {
            if device.key() != solana_program::system_program::id() && !force {
                return err!(ErrorCode::OrphanedDevices);
            }
        }

        // TODO: The group
        if group.deployments.len() > 0 && !force {
            return err!(ErrorCode::OrphanedDeployments);
        }

        let withdrawing_license_tokens = &mut ctx.accounts.withdrawing_license_tokens;
        let group_license_tokens = &mut ctx.accounts.group_license_tokens;

        let transfer_inst = Transfer {
            from: group_license_tokens.to_account_info(),
            to: withdrawing_license_tokens.to_account_info(),
            authority: group.to_account_info(),
        };

        let seeds = &[group.identifier.as_bytes(), b"work_group", &[group.bump]];

        let signer: &[&[&[u8]]] = &[&seeds[..]];
        let cpi_ctx = CpiContext::new_with_signer(
            ctx.accounts.token_program.to_account_info(),
            transfer_inst,
            signer,
        );
        token::transfer(cpi_ctx, group_license_tokens.amount)?;

        // CloseAccount{} on group_license_tokens
        Ok(())
    }

    pub fn register_device(ctx: Context<RegisterDevice>, device_authority: Pubkey) -> Result<()> {
        let work_group = &mut ctx.accounts.work_group;
        let group_license_tokens = &mut ctx.accounts.group_license_tokens;
        let devices_len = work_group.devices.len() as u64;
        if devices_len + 1 > group_license_tokens.amount * 10 {
            return err!(ErrorCode::InsufficentLicenseTokens);
        }
        let device = &mut ctx.accounts.device;

        device.work_group = work_group.key();
        device.device_authority = device_authority;
        work_group.devices.push(device.to_account_info().key());
        Ok(())
    }

    pub fn close_device(ctx: Context<CloseDevice>) -> Result<()> {
        let work_group = &mut ctx.accounts.work_group;
        let device_key = ctx.accounts.device.to_account_info().key;
        let index_in_group = work_group
            .devices
            .iter()
            .position(|x| *x == *device_key)
            .unwrap();
        // Currently networking depend on the index of the device in the group,
        // so don't delete it, just set the pubkey to system program ID.
        work_group.devices[index_in_group] = solana_program::system_program::ID;
        Ok(())
    }

    pub fn update_device(
        ctx: Context<UpdateDevice>,
        ipv4: [u8; 4],
        hostname: String,
        bump: u8,
        status: DeviceStatus,
    ) -> ProgramResult {
        ctx.accounts.device.ipv4 = ipv4;
        ctx.accounts.device.hostname = hostname;
        ctx.accounts.device.bump = bump;
        ctx.accounts.device.status = status;
        Ok(())
    }

    pub fn create_work_spec(
        ctx: Context<CreateWorkSpec>,
        spec_name: String,
        work_type: WorkType,
        url_or_contents: String,
        contents_sha256: String,
        metadata_url: String,
        mutable: bool,
    ) -> Result<()> {
        let spec = &mut ctx.accounts.spec;
        let work_group = &mut ctx.accounts.work_group;
        let clock = Clock::get().unwrap();
        spec.created_at = clock.unix_timestamp as u64;
        spec.modified_at = clock.unix_timestamp as u64;
        spec.name = spec_name;
        spec.url_or_contents = url_or_contents;
        spec.work_type = work_type;
        spec.mutable = mutable;
        spec.contents_sha256 = contents_sha256;
        spec.metadata_url = metadata_url;
        work_group.specs.push(spec.to_account_info().key());
        Ok(())
    }

    pub fn close_work_spec(ctx: Context<CloseWorkSpec>) -> Result<()> {
        let work_group = &mut ctx.accounts.work_group;
        let spec_key = ctx.accounts.spec.to_account_info().key;
        let index_in_group = work_group
            .specs
            .iter()
            .position(|x| *x == *spec_key)
            .unwrap();
        work_group.specs.remove(index_in_group);
        Ok(())
    }

    pub fn create_deployment(
        ctx: Context<CreateDeployment>,
        name: String,
        replicas: u8,
    ) -> ProgramResult {
        let deployment = &mut ctx.accounts.deployment;
        deployment.spec = ctx.accounts.spec.to_account_info().key();
        deployment.name = name;
        let deployment_mint = &mut ctx.accounts.deployment_mint;
        let deployment_tokens = &mut ctx.accounts.deployment_tokens;
        let work_group = &mut ctx.accounts.work_group;

        deployment.deployment_bump = *ctx.bumps.get("deployment").unwrap();
        deployment.mint_bump = *ctx.bumps.get("deployment_mint").unwrap();
        deployment.tokens_bump = *ctx.bumps.get("deployment_tokens").unwrap();
        deployment.replicas = replicas;

        let cpi_accounts = MintTo {
            mint: deployment_mint.to_account_info(),
            to: deployment_tokens.to_account_info(),
            authority: deployment.to_account_info(),
        };
        let cpi_program = ctx.accounts.token_program.to_account_info();
        let seeds = &[
            work_group.to_account_info().key.as_ref(),
            deployment.name.as_bytes(),
            b"deployment",
            &[deployment.deployment_bump],
        ];
        let signer: &[&[&[u8]]] = &[&seeds[..]];
        let cpi_ctx = CpiContext::new_with_signer(cpi_program, cpi_accounts, signer);
        token::mint_to(cpi_ctx, replicas as u64)?;
        work_group
            .deployments
            .push(deployment.to_account_info().key());

        Ok(())
    }

    pub fn close_deployment(ctx: Context<CloseDeployment>) -> Result<()> {
        let work_group = &mut ctx.accounts.work_group;
        let deployment_key = ctx.accounts.deployment.to_account_info().key;
        let index_in_group = work_group
            .deployments
            .iter()
            .position(|x| *x == *deployment_key)
            .unwrap();
        work_group.deployments.remove(index_in_group);

        let deployment_tokens = &ctx.accounts.deployment_tokens;
        let deployment_mint = &ctx.accounts.deployment_mint;
        let deployment = &ctx.accounts.deployment;

        let seeds = &[
            work_group.to_account_info().key.as_ref(),
            deployment.name.as_bytes(),
            b"deployment",
            &[deployment.deployment_bump],
        ];
        let signer: &[&[&[u8]]] = &[&seeds[..]];

        // First send all tokens remaining in deployment token
        // account back to the mint (must have balance of zero to close
        // token account)
        let burn_inst = Burn {
            from: deployment_tokens.to_account_info(),
            mint: deployment_mint.to_account_info(),
            authority: deployment.to_account_info(),
        };
        let cpi_ctx = CpiContext::new_with_signer(
            ctx.accounts.token_program.to_account_info(),
            burn_inst,
            signer,
        );
        token::burn(cpi_ctx, deployment_tokens.amount)?;

        // Next call CloseAccount on the token account
        let close_token_acct_inst = CloseAccount {
            account: ctx.accounts.deployment_tokens.to_account_info(),
            destination: ctx.accounts.group_authority.to_account_info(),
            authority: ctx.accounts.deployment.to_account_info(),
        };
        let cpi_ctx = CpiContext::new_with_signer(
            ctx.accounts.token_program.to_account_info(),
            close_token_acct_inst,
            signer,
        );
        token::close_account(cpi_ctx)?;

        Ok(())
    }

    pub fn schedule(ctx: Context<Schedule>, replicas: u8) -> Result<()> {
        let deployment = &mut ctx.accounts.deployment;
        let deployment_tokens = &mut ctx.accounts.deployment_tokens;
        let device_tokens = &mut ctx.accounts.device_tokens;
        let work_group = &ctx.accounts.work_group;

        if replicas as u64 > deployment_tokens.amount {
            return err!(ErrorCode::InsuffientReplicaTokens);
        }

        let transfer_inst = Transfer {
            from: deployment_tokens.to_account_info(),
            to: device_tokens.to_account_info(),
            authority: deployment.to_account_info(),
        };

        let seeds = &[
            work_group.to_account_info().key.as_ref(),
            deployment.name.as_bytes(),
            b"deployment",
            &[deployment.deployment_bump],
        ];
        let signer: &[&[&[u8]]] = &[&seeds[..]];
        let cpi_ctx = CpiContext::new_with_signer(
            ctx.accounts.token_program.to_account_info(),
            transfer_inst,
            signer,
        );
        token::transfer(cpi_ctx, replicas as u64)?;

        let approve_inst = Approve {
            to: deployment_tokens.to_account_info(),
            delegate: work_group.to_account_info(),
            authority: deployment.to_account_info(),
        };
        let cpi_ctx = CpiContext::new_with_signer(
            ctx.accounts.token_program.to_account_info(),
            approve_inst,
            signer,
        );

        // approve group as delegate on token account, that way
        // it can unschedule the scheduled token(s)
        token::approve(cpi_ctx, replicas as u64)?;
        Ok(())
    }
}

#[derive(Clone, AnchorSerialize, AnchorDeserialize, PartialEq)]
pub enum WorkType {
    DockerCompose,
    // TODO: Could have more drivers like
    // Kubernetes, Nomad, etc.
}

impl Default for WorkType {
    fn default() -> Self {
        WorkType::DockerCompose
    }
}

#[account]
#[derive(Default)]
pub struct WorkSpec {
    name: String,
    work_type: WorkType,
    created_at: u64,
    modified_at: u64,
    url_or_contents: String,
    contents_sha256: String,
    metadata_url: String,
    mutable: bool,
    // bump: u8, // ?
}

#[derive(Clone, Copy, AnchorSerialize, AnchorDeserialize)]
pub enum DeviceStatus {
    RegistrationRequested,
    Registered,
    Delinquent,
    Cordoned,
}

impl Default for DeviceStatus {
    fn default() -> Self {
        DeviceStatus::RegistrationRequested
    }
}

#[account]
#[derive(Default)]
pub struct Device {
    pub ipv4: [u8; 4],
    pub hostname: String,
    pub bump: u8,
    pub status: DeviceStatus,
    pub device_authority: Pubkey,
    pub work_group: Pubkey,
}

#[account]
#[derive(Default)]
pub struct WorkGroup {
    // discriminator: 8 bytes
    pub bump: u8,
    pub group_authority: Pubkey,

    // todo: these might make more sense as like
    //       a Set or some other type of data structure
    //       cause I'm assuming we might want to delete
    //       at some point
    pub specs: Vec<Pubkey>,
    pub devices: Vec<Pubkey>,
    pub deployments: Vec<Pubkey>,

    pub name: String,
    pub identifier: String,
    pub signal_server_url: String,
}

#[derive(Accounts)]
pub struct CloseWorkGroup<'info> {
    #[account(mut)]
    pub group_authority: Signer<'info>,

    #[account(mut,
        close = group_authority,
        seeds = [group.identifier.as_bytes(), b"work_group"],
        bump,
        has_one = group_authority,
    )]
    pub group: Box<Account<'info, WorkGroup>>,

    #[account(address = worknet_license_token::ID)]
    pub license_mint: Box<Account<'info, Mint>>,

    #[account(mut,
        token::mint = license_mint,
        token::authority = group,
        seeds = [group.key().as_ref(), b"license_tokens"],
        bump,
    )]
    pub group_license_tokens: Box<Account<'info, TokenAccount>>,

    #[account(mut,
        token::mint = license_mint,
        token::authority = group_authority,
    )]
    pub withdrawing_license_tokens: Box<Account<'info, TokenAccount>>,

    pub rent: Sysvar<'info, Rent>,
    pub token_program: Program<'info, Token>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
#[instruction(name: String, identifier: String)]
pub struct CreateWorkGroup<'info> {
    #[account(mut)]
    pub group_authority: Signer<'info>,

    #[account(init,
        seeds = [identifier.as_bytes(), b"work_group"],
        bump,
        payer = group_authority,
        space = 1028
    )]
    pub group: Box<Account<'info, WorkGroup>>,

    #[account(address = worknet_license_token::ID)]
    pub license_mint: Box<Account<'info, Mint>>,

    #[account(init,
        token::mint = license_mint,
        token::authority = group,

        // TODO: Make more sense to use group or authority here?
        seeds = [group.key().as_ref(), b"license_tokens"],
        bump,
        payer = group_authority
    )]
    pub group_license_tokens: Box<Account<'info, TokenAccount>>,

    #[account(mut,
        token::mint = license_mint,
        token::authority = group_authority,
    )]
    pub depositing_license_tokens: Box<Account<'info, TokenAccount>>,

    pub rent: Sysvar<'info, Rent>,
    pub token_program: Program<'info, Token>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
#[instruction(spec_name: String)]
pub struct CreateWorkSpec<'info> {
    #[account(mut)]
    pub group_authority: Signer<'info>,

    #[account(init,
        seeds = [work_group.key().as_ref(), spec_name.as_bytes(), b"spec"],
        bump,
        payer = group_authority,
        space = 512
    )]
    pub spec: Box<Account<'info, WorkSpec>>,

    #[account(mut,
        has_one = group_authority,
        seeds = [
            work_group.identifier.as_bytes(),
            b"work_group",
        ],
        bump,
    )]
    pub work_group: Box<Account<'info, WorkGroup>>,

    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct CloseWorkSpec<'info> {
    #[account(mut)]
    pub group_authority: Signer<'info>,

    #[account(mut,
        close = group_authority,
        seeds = [work_group.key().as_ref(), spec.name.as_bytes(), b"spec"],
        bump,
    )]
    pub spec: Box<Account<'info, WorkSpec>>,

    #[account(mut,
        has_one = group_authority,
        seeds = [
            work_group.identifier.as_bytes(),
            b"work_group",
        ],
        bump,
    )]
    pub work_group: Box<Account<'info, WorkGroup>>,

    pub system_program: Program<'info, System>,
}

#[derive(Clone, AnchorSerialize, AnchorDeserialize)]
pub enum DeploymentArgType {
    String,
    Number,
}

impl Default for DeploymentArgType {
    fn default() -> Self {
        DeploymentArgType::String
    }
}

#[derive(Clone, AnchorSerialize, AnchorDeserialize)]
pub struct DeploymentArg {
    pub arg_name: String,
    pub arg_value: String,
    pub arg_type: DeploymentArgType,
}

#[account]
#[derive(Default)]
pub struct Deployment {
    pub spec: Pubkey,
    pub name: String,
    pub args: Vec<DeploymentArg>,
    pub replicas: u8,
    pub self_bump: u8,
    pub mint_bump: u8,
    pub tokens_bump: u8,
    pub deployment_bump: u8,
}

#[derive(Accounts)]
#[instruction(name: String)]
pub struct CreateDeployment<'info> {
    #[account(mut)]
    pub group_authority: Signer<'info>,

    #[account(init,
        seeds = [work_group.key().as_ref(), name.as_bytes(), b"deployment"],
        bump,
        payer = group_authority,
        space = 128,
    )]
    pub deployment: Box<Account<'info, Deployment>>,

    // todo: make it so you can look this up by name
    pub spec: Box<Account<'info, WorkSpec>>,

    // Don't need fractional token quantities, just int,
    // so decimals is set to 0.
    #[account(init,
        mint::decimals = 0,
        mint::authority = deployment,
        seeds = [deployment.key().as_ref(), b"deployment_mint"],
        bump,
        payer = group_authority
    )]
    pub deployment_mint: Box<Account<'info, Mint>>,

    #[account(init,
        token::mint = deployment_mint,
        token::authority = deployment,
        seeds = [deployment.key().as_ref(), b"deployment_tokens"],
        bump,
        payer = group_authority
    )]
    pub deployment_tokens: Box<Account<'info, TokenAccount>>,

    #[account(mut,
        has_one = group_authority,
        seeds = [
            work_group.identifier.as_bytes(),
            b"work_group",
        ],
        bump,
    )]
    pub work_group: Box<Account<'info, WorkGroup>>,

    pub system_program: Program<'info, System>,
    pub token_program: Program<'info, Token>,
    pub rent: Sysvar<'info, Rent>,
}

#[derive(Accounts)]
pub struct CloseDeployment<'info> {
    #[account(mut)]
    pub group_authority: Signer<'info>,

    #[account(mut,
        close = group_authority,
        seeds = [work_group.key().as_ref(), deployment.name.as_bytes(), b"deployment"],
        bump,
    )]
    pub deployment: Box<Account<'info, Deployment>>,

    #[account(mut,
        mint::authority = deployment.key(),
        seeds = [deployment.key().as_ref(), b"deployment_mint"],
        bump,
    )]
    pub deployment_mint: Box<Account<'info, Mint>>,

    #[account(mut,
        token::mint = deployment_mint,
        token::authority = deployment,
        seeds = [deployment.key().as_ref(), b"deployment_tokens"],
        bump,
    )]
    pub deployment_tokens: Box<Account<'info, TokenAccount>>,

    #[account(mut,
        has_one = group_authority,
        seeds = [
            work_group.identifier.as_bytes(),
            b"work_group",
        ],
        bump,
    )]
    pub work_group: Box<Account<'info, WorkGroup>>,

    pub system_program: Program<'info, System>,
    pub token_program: Program<'info, Token>,
    pub rent: Sysvar<'info, Rent>,
}

#[derive(Accounts)]
pub struct UpdateDevice<'info> {
    pub device_authority: Signer<'info>,

    // TODO: daolet should be start with
    // group authority flag and UpdateDevice
    // should receive this authority and validate
    // the authority is correct.
    #[account(mut,
        has_one = device_authority,
        seeds = [device.device_authority.key().as_ref()],
        bump,
    )]
    pub device: Box<Account<'info, Device>>,
}

#[derive(Accounts)]
#[instruction(device_authority: Pubkey)]
pub struct RegisterDevice<'info> {
    #[account(mut)]
    pub group_authority: Signer<'info>,

    #[account(init,
        space = 128,
        seeds = [device_authority.as_ref()],
        bump,
        payer = group_authority,
    )]
    pub device: Box<Account<'info, Device>>,

    #[account(address = worknet_license_token::ID)]
    pub license_mint: Box<Account<'info, Mint>>,

    #[account(
        token::mint = license_mint,
        token::authority = work_group,
        seeds = [work_group.key().as_ref(), b"license_tokens"],
        bump,
    )]
    pub group_license_tokens: Box<Account<'info, TokenAccount>>,

    #[account(mut,
        has_one = group_authority,
        seeds = [
            work_group.identifier.as_bytes(),
            b"work_group",
        ],
        bump,
    )]
    pub work_group: Box<Account<'info, WorkGroup>>,

    pub system_program: Program<'info, System>,
    pub token_program: Program<'info, Token>,
    // TODO: TTL?
}

#[derive(Accounts)]
pub struct CloseDevice<'info> {
    #[account(mut)]
    pub group_authority: Signer<'info>,

    #[account(mut,
        has_one = work_group,
        close = group_authority,
    )]
    pub device: Box<Account<'info, Device>>,

    #[account(mut,
        has_one = group_authority,
        seeds = [
            work_group.identifier.as_bytes(),
            b"work_group",
        ],
        bump,
    )]
    pub work_group: Box<Account<'info, WorkGroup>>,

    pub system_program: Program<'info, System>,
    pub token_program: Program<'info, Token>,
    // TODO: TTL?
}

#[derive(Accounts)]
pub struct Schedule<'info> {
    #[account(mut)]
    pub group_authority: Signer<'info>,

    #[account(
        has_one = group_authority,
        seeds = [
            work_group.identifier.as_bytes(),
            b"work_group",
        ],
        bump,
    )]
    pub work_group: Box<Account<'info, WorkGroup>>,

    #[account(
        seeds = [work_group.key().as_ref(), deployment.name.as_bytes(), b"deployment"],
        bump,
    )]
    pub deployment: Box<Account<'info, Deployment>>,

    #[account(
        mint::authority = deployment.key(),
        seeds = [deployment.key().as_ref(), b"deployment_mint"],
        bump,
    )]
    pub deployment_mint: Box<Account<'info, Mint>>,

    #[account(mut,
        token::mint = deployment_mint,
        token::authority = deployment,
        seeds = [deployment.key().as_ref(), b"deployment_tokens"],
        bump,
    )]
    pub deployment_tokens: Box<Account<'info, TokenAccount>>,

    #[account(
        seeds = [device.device_authority.key().as_ref()],
        bump,
    )]
    pub device: Box<Account<'info, Device>>,

    /// CHECK: just used to pass through key for token account owner
    pub device_authority: UncheckedAccount<'info>,

    // TODO: init_if_needed is considered potentially
    // vulnerable to re-initialization attacks. In theory,
    // this might be better served as something like:
    //
    // (1) On client side, check if token account exists
    // (2) If not, call an instruction to create it
    // (3) Use 'mut' on this method not 'init'
    //
    // docs here:
    // https://docs.rs/anchor-lang/latest/anchor_lang/derive.Accounts.html
    #[account(init_if_needed,
        token::mint = deployment_mint,
        token::authority = device_authority,
        seeds = [
            device_authority.key().as_ref(),
            deployment.key().as_ref(),
            b"device_tokens"
        ],
        bump,
        payer = group_authority,
    )]
    pub device_tokens: Box<Account<'info, TokenAccount>>,

    pub system_program: Program<'info, System>,
    pub token_program: Program<'info, Token>,
    pub rent: Sysvar<'info, Rent>,
}
