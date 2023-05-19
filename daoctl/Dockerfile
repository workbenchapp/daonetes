
# NOTE: this Dockerfile is for use by goreleaser

FROM ubuntu:latest

RUN apt-get update \
    && apt-get install -yq apt-transport-https ca-certificates curl gnupg-agent software-properties-common

RUN curl https://download.docker.com/linux/static/stable/$(uname -m)/docker-20.10.9.tgz --output docker.tgz \
    && tar xzvf docker.tgz \
    && mv docker/* /usr/local/bin \
    && rm -r docker/ docker.tgz \
    && curl -L "https://github.com/docker/compose/releases/download/v2.9.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose \
    && chmod +x /usr/local/bin/docker-compose

WORKDIR /var/lib/worknet
RUN apt-get install -yq wireguard iproute2 inetutils-ping

# ip link add dev wg0 type wireguard
COPY worknet /usr/bin/
ENTRYPOINT [ "/usr/bin/worknet" ]
CMD [ "daolet" ]