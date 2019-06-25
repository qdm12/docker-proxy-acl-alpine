ARG ALPINE_VERSION=3.10
ARG GO_VERSION=1.12.6

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
ARG BINCOMPRESS
RUN apk --update add git build-base upx
RUN go get -u -v golang.org/x/vgo
WORKDIR /tmp/gobuild
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
# RUN go test -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-s -w" -o app .
RUN [ "${BINCOMPRESS}" == "" ] || (upx -v --best --ultra-brute --overlay=strip app && upx -t app)

FROM scratch
ARG BUILD_DATE
ARG VCS_REF
LABEL org.label-schema.schema-version="1.0.0-rc1" \
    maintainer="quentin.mcgaw@gmail.com" \
    org.label-schema.build-date=$BUILD_DATE \
    org.label-schema.vcs-ref=$VCS_REF \
    org.label-schema.vcs-url="https://github.com/qdm12/docker-proxy" \
    org.label-schema.url="https://github.com/qdm12/docker-proxy" \
    org.label-schema.vcs-description="Lightweight container running a restricted Docker unix socket proxy" \
    org.label-schema.vcs-usage="https://github.com/qdm12/docker-proxy/blob/master/README.md#setup" \
    org.label-schema.docker.cmd="docker run -d qmcgaw/docker-proxy-acl-alpine" \
    org.label-schema.docker.cmd.devel="docker run -it --rm qmcgaw/docker-proxy-acl-alpine" \
    org.label-schema.docker.params="" \
    org.label-schema.version="" \
    image-size="5.82MB" \
    ram-usage="10MB" \
    cpu-usage="Low"
ENTRYPOINT ["/dockerproxy"]
# HEALTHCHECK --interval=300s --timeout=5s --start-period=5s --retries=1 CMD ["/healthcheck"]   
# USER 1000
CMD -a containers
COPY --from=builder /tmp/gobuild/app /dockerproxy