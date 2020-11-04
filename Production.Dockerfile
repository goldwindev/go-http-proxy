FROM golang:alpine as builder

ENV GO111MODULE=on

RUN \
  apk update \
  && apk add --no-cache make wget curl gnupg git

WORKDIR /go/src/go-secure-proxy
COPY . .

RUN make build

FROM alpine:latest

WORKDIR /tmp/
COPY --from=builder /go/src/go-secure-proxy/bin/go-secure-proxy-linux-amd64 /tmp/

CMD ["/tmp/go-secure-proxy-linux-amd64"]
