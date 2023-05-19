# daoctl

the easiest way to work on `daoctl` is to install the latest golang, and then run:

```
go run main.go <params>
```

## goreleaser

### local builds

ideally we'll always use https://goreleaser.com/install/#running-with-docker

local builds can be done with

```
goreleaser release  --skip-publish  --snapshot --rm-dist
```

### Building a release

Release builds are done using GitHub Actions. To build a new release, create a new annotated tag and push it to the repo - GitHub will take care of the rest.:

```
sven@p1:~/src/worknet$ git tag -a v0.0.6 -m "Fixed some of the issues in the agent"
sven@p1:~/src/worknet$ git push origin v0.0.6
Enumerating objects: 1, done.
Counting objects: 100% (1/1), done.
Writing objects: 100% (1/1), 184 bytes | 184.00 KiB/s, done.
Total 1 (delta 0), reused 0 (delta 0), pack-reused 0
To github.com:workbenchapp/worknet
 * [new tag]         v0.0.6 -> v0.0.6
```

### install / update

```
go install github.com/goreleaser/goreleaser@latest
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
```

## RUN the daolet container

```
docker run --rm -it -v /run/docker.sock:/run/docker.sock --publish 51820:51820 -v daolet-cfg:/root/.config  --cap-add=NET_ADMIN   --cap-add=SYS_MODULE daonetes/worknet
```
