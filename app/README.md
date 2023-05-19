# DAOnetes DAPP

1. Should let the user open an existing Solana wallet, or https://magic.link/ create one
2. then allows them to list all the Realms they're in, or create one
3. select a realm you're in and you can list users, devices and specs
4. propose new users, propose new specs to deploy

## Run the DAPP

1. `npm install`
2. `npm run dev`
3. browse to [http://localhost:3000](http://localhost:3000)

## Project start..

This is an app bootstrapped according to the [init.tips](https://init.tips) stack, also known as the T3-Stack.

## Project Ideals

We want the DAPP to be statically built, installable on a mobile/browser as a PWA, so that it only needs to be able to connect to its Solana validator network. It should be possible to use a private, internal validator network.

- `react-bootstrap` for the UI
- offload all keysigning to the external wallet-adapter
- default to showing on-chain info only
- but _if_ we detect that we have access to one of the in-Realm nodes, we should be able to give a more live view of running state.
- need to make a cross-realm reward system to encourage sharing of spec knowledge
