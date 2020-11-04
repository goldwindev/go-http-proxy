FROM golang:alpine as builder

RUN \
  apk update \
  && apk add --no-cache make wget curl gnupg git

WORKDIR /go/src/go-secure-proxy
COPY . .

RUN make build

FROM builder

WORKDIR /tmp/
COPY ./bin/go-secure-proxy-linux-amd64 /tmp/

CMD ["/tmp/go-secure-proxy-linux-amd64"]
