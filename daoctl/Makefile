.PHONY: signal-server

signal-server:
	docker build --pull -t daonetes/signal-server:latest -f Dockerfile.signal-server .

push-signal-server:
	docker push daonetes/signal-server:latest

agent:
	goreleaser release  --skip-publish  --rm-dist  --snapshot
