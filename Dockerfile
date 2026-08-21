FROM golang:1.26-alpine3.22 AS base

ENV CGO_ENABLED=0
ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git build-base

WORKDIR /go/src/video-orchestrator
COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM base AS dev

RUN go install github.com/M2G/modd/cmd/modd@v0.1.2
RUN mv "$(go env GOPATH)/bin/modd" /usr/local/bin/modd

EXPOSE 8181
ENTRYPOINT ["modd"]
CMD ["-f", "configuration/modd/modd.conf"]


FROM base AS builder

RUN make build


FROM alpine:latest AS release

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/video-orchestrator/bin/video-orchestrator /video-orchestrator

EXPOSE 8181
ENTRYPOINT ["/video-orchestrator"]