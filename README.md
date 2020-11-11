# Go HTTP Proxy
Reverse proxy written in go. It proxies requests from `PORT` to `PROXY_URL`.

## Build
`docker build -t go-http-proxy -f Production.Dockerfile .`

## Run
`docker run --env PORT=80 --env PROXY_URL=https://example.com --network host go-http-proxy:latest`
