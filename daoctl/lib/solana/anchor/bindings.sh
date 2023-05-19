#!/bin/bash

set -euxo pipefail

anchor build
anchor idl parse -f ../../../../programs/worknet/src/lib.rs -o worknet_idl.json
anchor-go --src=worknet_idl.json

gen_goki() {
	git clone \
		--depth 1 \
		git@github.com:GokiProtocol/goki.git \
		/tmp/goki
	cd /tmp/goki
	./scripts/parse-idls.sh
	cd -
	cp /tmp/goki/artifacts/idl/smart_wallet.json \
		goki_smart_wallet.json
	rm -rf /tmp/goki

	# rename cause smart_wallet package name
	# would look clunky in Go
	sed -i \
		's/"name": "smart_wallet"/"name": "smartwallet"/' \
		goki_smart_wallet.json
	rm -rf generated/smartwallet
	anchor-go --src=goki_smart_wallet.json
}

if [[ -n "${ANCHOR_GEN_ALL:-}" ]]
then
	gen_goki
fi
