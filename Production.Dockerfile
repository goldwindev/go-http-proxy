FROM golang:alpine as builder

ENV GO111MODULE=on

RUN \
  apk update \
  && apk add --no-cache make wget curl gnupg git

WORKDIR /go/src/go-http-proxy
COPY . .

RUN make linux

FROM arm64v8/alpine:latest

WORKDIR /tmp/
COPY --from=builder /go/src/go-http-proxy/bin/go-http-proxy-linux-arm64 /tmp/

CMD ["/tmp/go-http-proxy-linux-arm64"]
