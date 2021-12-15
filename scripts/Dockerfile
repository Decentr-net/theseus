ARG ALPINE_VERSION=3.10
ARG GOLANG_VERSION=1.16.9

FROM golang:${GOLANG_VERSION}
WORKDIR /go/src/github.com/Decentr-net/theseus/
COPY . .
RUN make linux

FROM alpine:${ALPINE_VERSION}
RUN apk update && apk add ca-certificates
COPY --from=0 /go/src/github.com/Decentr-net/theseus/build/theseus-linux-amd64 /theseusd
COPY --from=0 /go/src/github.com/Decentr-net/theseus/build/sync-linux-amd64 /syncd
COPY static /static
COPY scripts/migrations /migrations
ENTRYPOINT [ "/theseusd" ]
