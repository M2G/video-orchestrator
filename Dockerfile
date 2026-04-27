FROM        golang:1.26.2-alpine3.23 as base
ENV         GO111MODULE     on

RUN         apk -u add git openssh build-base
WORKDIR     /go/src/github.com/video-orchestrator
ADD         .   /go/src/github.com/video-orchestrator
COPY        scripts/dep .
RUN         chmod +x dep; ./dep

FROM        base as dev
EXPOSE      80 443 43554
ADD         .  /go/src/github.com/video-orchestrator
WORKDIR     /go/src/github.com/video-orchestrator
RUN         wget https://github.com/cortesi/modd/releases/download/v0.8/modd-0.8-linux64.tgz
RUN         tar -xzf modd-0.8-linux64.tgz
RUN         mv ./modd-0.8-linux64/modd /usr/local/bin
ENTRYPOINT  ["modd"]
CMD         ["-f", "configuration/modd/modd.conf"]

FROM        base as builder
ADD         .   /go/src/github.com/video-orchestrator
WORKDIR     /go/src/github.com/video-orchestrator
RUN         chmod +x scripts/dep; ./scripts/dep
RUN         make build

FROM        alpine:latest as release
EXPOSE      80 443 43554
RUN         apk -u add ca-certificates
COPY        --from=builder /go/src/github.com/video-orchestrator/bin/video-orchestrator video-orchestrator
ENTRYPOINT  [ "/video-orchestrator" ]
