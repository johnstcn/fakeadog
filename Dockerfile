FROM golang:alpine as builder
WORKDIR /go/src/github.com/johnstcn/fakeadog
ADD . /go/src/github.com/johnstcn/fakeadog
RUN set -x && \
    apk add -q --update && \
    apk add -q curl git
RUN set -x && \
    curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && \
    chmod +x /usr/local/bin/dep
RUN set -x && \
    go version && \
    go env && \
    dep ensure -v && \
    CGO_ENABLED=0 go test -v ./... && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo github.com/johnstcn/fakeadog/cmd/fakeadog

FROM alpine:latest
MAINTAINER Cian Johnston <public@cianjohnston.ie>

ENV HOST 0.0.0.0
ENV PORT 8125
EXPOSE 8125/udp

WORKDIR /root
COPY --from=builder /go/src/github.com/johnstcn/fakeadog/fakeadog .

CMD ["./fakeadog"]
