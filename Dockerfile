FROM alpine
MAINTAINER James Tancock <james.tancock@momentumft.co.uk>

ENV GOPATH /go
ENV REPOSRC /go/src/github.com/momentumft/consul-gc

RUN apk update && apk add ca-certificates \
    && rm -rf /var/cache/apk/*

COPY . $REPOSRC

RUN packages='go git gcc libc-dev libgcc' \
    && apk update \
    && apk add $packages \
    && cd $REPOSRC \
    && go get -v ./... \
    && go build -x -o /usr/bin/consul-gc . \
    && apk del $packages \
    && rm -rf $GOPATH \
    && rm -rf /var/cache/apl/*

ENTRYPOINT ["/usr/bin/consul-gc"]
