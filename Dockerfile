FROM golang:alpine AS builder
RUN apk --update add git build-base upx
WORKDIR /go/src/app
COPY vendor/manifest vendor/manifest
COPY vendor/github.com/gorilla/context/*.go vendor/github.com/gorilla/context/
COPY vendor/github.com/gorilla/mux/*.go vendor/github.com/gorilla/mux/
COPY vendor/github.com/namsral/flag/*.go vendor/github.com/namsral/flag/
COPY docker-proxy-acl.go ./docker-proxy-acl.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -installsuffix cgo -o proxy . && \
    upx -v --best --overlay=strip proxy && \
    upx -t proxy

FROM alpine:3.8
LABEL maintainer="quentin.mcgaw@gmail.com" \
      description="Lightweight container running a restricted Docker unix socket proxy" \
      download="???MB" \
      size="6MB" \
      ram="???MB" \
      cpu_usage="Very low to low" \
      github="https://github.com/qdm12/docker-proxy-acl"
ENV OPTIONS -a containers
COPY --from=builder /go/src/app/proxy /proxy
ENTRYPOINT /proxy "$OPTIONS"
