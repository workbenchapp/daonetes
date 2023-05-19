# Run

You may want to tell the solana cli to use your local validator, or devnet/testnet..

`solana config set --url http://localhost:8899`

You need to have an account with sufficient SOL (2.05891416 SOL) to deploy lib.rs into the validator

```
yarn
solana-keygen new -o ~/.config/solana/id.json   <<<<USE Solana-Workbench and use its wallet account..., then use `solana-keygen recover -o ~/.config/solana/id.json  ASK` and enter the mnemonic phrase for the ElectronAppStorageKeypair private key in `~/.config/SolanaWorkbench/electron-cfg.json`>>>>
solana airdrop 666 $(solana address)
solana balance $(solana address)
anchor build
anchor deploy
```

To run the Anchor tests, the secret key JSON for a wallet that holds license
tokens on devnet must be set in the `WORKNET_TEST_PAYER` environment variable.

e.g.:

```
WORKNET_TEST_PAYER=$(cat ~/worknet_test_payer.json) anchor test
```

Anchor clones accounts for the license token mint and a license token account
(i.e., the token account held by the key inserted above) from devnet when it
runs its test suite.

> > NOTE: at this point, you can "Add Account" in Solana workbench (tho it didn't show in my live view?)

# Using Daonetes

## Deploy daonetes program

The daonetes (worknet) program can function independently, but the CLI ~~depends
on [Goki](https://goki.so) for smart wallet support (multisig) so that multiple
users can access the same cluster. That requires Goki program to be present for
local development.~~ 2023 edit: The Macaroni brothers are fuckboys, therefore we 
can't have nice things. If designing now, we would probably use [Squads](https://squads.so/),
however there is also support for doing group based things, albeit in a clunky way, with the built in
SPL Governance Program (Realms) supported by the front end. Still, it's a bummer the 
git-like workflow from before that used Goki is no longer supported.

[Amman](https://github.com/metaplex-foundation/amman) is used
for dependencies. It can also start a validator with the worknet program .so once that
program has been built.

To install Amman:

```
$ npm install -g @metaplex-foundation/amman
```

Build the worknet program and starting a validator from worknet repo root:

```
$ anchor build
$ amman start
```

This requires internet access on the first run. Amman will clone Goki locally
from mainnet.

## Worknet Architecture

![](/daoctl/docs/worknet_architecture.png?raw=true)

## Creating a cluster and adding machines

First, you need a named 'work group' (cluster). Default name is `default`.

```
$ daoctl group init
```

Then you need to register the worker device(s) with the pubkey of their node. `daoctl agent` on the worker node (pre-registration) will give you the command to do so.

```
$ daoctl agent
Using existing /Users/nathanleclaire/.config/WorkNet/Wallet.json public key: ERMyKD5bKFRrpRcTKwuaat1HbjixfgzNJv2bCRodARVC
Looking for PDA C98bt2yQ2XK7u8dKUkAzJ4yEGUL2z28JTiaH5vrmtJyC
daoctl: error: no PDA found. Must register device:
               daoctl device register ERMyKD5bKFRrpRcTKwuaat1HbjixfgzNJv2bCRodARVC
```

then:

```
$ daoctl device register <nodekey>
```

then from the worker device run the agent:

```
$ daoctl agent
```

(note: I just noticed [a bug](https://github.com/workbenchapp/worknet/issues/16) with the funding for the worker device, I will work on fixing it, but for now you can airdrop to the device - e.g. `solana -k /Users/nathanleclaire/.config/WorkNet/Wallet.json airdrop 1`)

This will not exit because it will begin polling for work.

You should be able to verify the device is Registered once the transaction completes:

```
$ daoctl device
HOSTNAME                    IPV4          STATUS       AUTHORITY                                      PDA
Nathans-MacBook-Pro.local   10.1.44.154   Registered   ERMyKD5bKFRrpRcTKwuaat1HbjixfgzNJv2bCRodARVC   C98bt2yQ2XK7u8dKUkAzJ4yEGUL2z28JTiaH5vrmtJyC
```

in this output --

PDA refers to the device data account
AUTHORITY refers to the device's unique key

## Adding specs and deploying them

Now, to deploy. From the 'control node', first you must create and name a spec. There is limited support for understanding Docker Compose files (volumes, port mappings, container image). There is an example nginx compose file in the repo:

```
$ daoctl spec register -f ./daoctl/test/nginx/docker-compose.yml nginx
...

$ daoctl spec
NAME      CONTAINERS   KEY
nginx     web          Cevnyos8KwyvZJTPU1oFhDYkSUwLqyCibbq9Zg2G8UYs
```

You then use the spec PDA (KEY) to make a deploy, a request that work should be scheduled somewhere:

```
$ daoctl deploy create nginx_test_run Cevnyos8KwyvZJTPU1oFhDYkSUwLqyCibbq9Zg2G8UYs

...
$ daoctl deploy
NAME             SPEC KEY                                       KEY                                            REPLICAS
nginx_test_run   Cevnyos8KwyvZJTPU1oFhDYkSUwLqyCibbq9Zg2G8UYs   27F6bPQ6cS6AY8Ffr5UKB26bvn8L99aMZD9Bkq1aSLae   0
```

Creating a deploy will create a set of tokens. A token can be sent to a device's key to ensure it runs that replica. This is performed using the `daoctl deploy schedule` command.

```
$ daoctl device
HOSTNAME                    IPV4          STATUS       AUTHORITY                                      PDA
Nathans-MacBook-Pro.local   10.1.44.154   Registered   ERMyKD5bKFRrpRcTKwuaat1HbjixfgzNJv2bCRodARVC   C98bt2yQ2XK7u8dKUkAzJ4yEGUL2z28JTiaH5vrmtJyC

$ daoctl deploy schedule nginx_test_run ERMyKD5bKFRrpRcTKwuaat1HbjixfgzNJv2bCRodARVC
```

You should now see output in the `daoctl agent`'s logs that the container has been detected in the token account and the container is running.

```
$ docker ps
CONTAINER ID   IMAGE                            COMMAND                  CREATED              STATUS              PORTS                                                                                                                      NAMES
ec21046eb72c   nginx                            "/docker-entrypoint.…"   About a minute ago   Up About a minute   0.0.0.0:8080->80/tcp                                                                                                       tender_cannon
...

$ docker inspect -f '{{ .Config.Labels }}' ec21046eb72c
map[maintainer:NGINX Docker Maintainers <docker-maint@nginx.com> worknet.deployment_id:27F6bPQ6cS6AY8Ffr5UKB26bvn8L99aMZD9Bkq1aSLae]
```

## Managing smart wallet users

Multiple users can be added to the smart wallet for the work group. (Currently,
the wallet is created with 1/N approval threshold so each user effectively has
root access)

To view current users' pubkeys:

```
$ daoctl group
PUBKEY                                         BASE
HEwe5nVn4y2gV4tCQ3hZ6a3QxB9a72NyXstWxBk4q5ct   *
7JqbFTpZ5XSnfYTP9uvPso6Sx5NcqCbv2Q3JUoCrzxui
```

The "base" key used to originally create the smart wallet is marked with an
asterisk).

To add a user:

```
$ daoctl group add <pubkey>
```

To remove a user:

```
$ daoctl group rm <pubkey>
```

## Browsing in explorer

If you have depoyed with `anchor test`, you can deploy the IDL so you can see the decoded accounts on-chain using the Solana explorer:

```
$ anchor idl init -f ./target/idl/worknet.json EdUCoDdRnT5HsQ2Ejy3TWMTQP8iUyMQB4WzoNh45pNX9
```

## License Mints and Token Accounts

DAOnetes uses license tokens to gate access to the program. The table below
summarizes which mints + token accounts are involved and how they're accessed
in each environment.

| Environment              | Mint Authority                                                                                                                                                           | Mint Address                                                                                                                           | Token Account Address                                                                                                                                      | Provisioning                                                                                                                                                                                                                                                                                                                    |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| devnet                   | [`EYkBZzBiGAWE5PsTyqydSceKdYCAqyCJDAvEoaZhoJ1G`](https://solscan.io/account/EYkBZzBiGAWE5PsTyqydSceKdYCAqyCJDAvEoaZhoJ1G?cluster=devnet) (Nathan's local testing wallet) | [`Ew5hokTuULRDsgnhKThGv3nrw3RPjiHASQZNcRNTHJ9Z`](https://solscan.io/token/Ew5hokTuULRDsgnhKThGv3nrw3RPjiHASQZNcRNTHJ9Z?cluster=devnet) | Depends on key. We have one [in the DAO](https://app.realms.today/dao/4R6mHB5hirj348JUyYcGc4iZNYoTm7udhQRtL9VebUEp/treasury?cluster=devnet), `3AyoD…QNNkN` | The mint was provisioned using Nathan's local development key and many tokens were minted then transferred to our devnet Realms DAO. Medium term, this should be migrated to a multisig of some kind. The deployed program knows this mint ID due to having the `local-license-mint` feature off in the program's `Cargo.toml`. |
| localhost + Anchor tests | `G5w7ic2s5L39mChqEKwXbHGL1mXduRGofECWcnQGRYcG` (local testing key in `daoctl/test/key.json`)                                                                             | `3CrKoTYzfbeenzQmhXMQzHM929kioTvsr1JtDSX9uET5`                                                                                         | `9RNxHPFffWYcnAf2BN521VzUmCFhXNGoMKm1BdrZ3JqE`                                                                                                             | The mint was provisioned on devnet using the test key checked in to the repo. The Amman config, and Anchor tests, pull this mint and pre-provisioned token account for the testing key down from devnet. The program also has the `local-license-mint` feature on by default when compiled to hardcode this testing mint.       |

## Setup your dev env

(I'm writing this on Linux first, as that's where I am)

This will get you nodejs, rust, anchor, amman.

1 install solana-workbench devtools: `curl --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/workbenchapp/solana-workbench/main/bin/setup.sh | sh`
2 rust: `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh`
3 metaplex amman: `npm install -g @metaplex-foundation/amman`
4 golang: `sudo snap install go --classic`
`go install github.com/goreleaser/goreleaser@latest`
`go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest`
argh! add `~/go/bin/` to your path...

## Deploy to devnet

(need to have correct keys available according to `Anchor.toml` and declared IDs)

```
$ anchor build -- --no-default-features && anchor --provider.cluster d upgrade --program-id EdUCoDdRnT5HsQ2Ejy3TWMTQP8iUyMQB4WzoNh45pNX9 target/deploy/worknet.so
```

(this assumes default program ID of `EdUCoDdRnT5HsQ2Ejy3TWMTQP8iUyMQB4WzoNh45pNX9`, to create a new program ID / deployment, just `anchor --provider.cluster d deploy` will suffice)

make sure to upgrade the IDL:

```
$ anchor --provider.cluster d idl upgrade -f target/idl/worknet.json EdUCoDdRnT5HsQ2Ejy3TWMTQP8iUyMQB4WzoNh45pNX9
```
