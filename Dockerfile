ARG ALPINE_VERSION=3.8
ARG GO_VERSION=1.11.2

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
RUN apk --update add git build-base upx
RUN go get -u -v golang.org/x/vgo
WORKDIR /tmp/gobuild

FROM scratch AS final
ARG BUILD_DATE
ARG VCS_REF
LABEL org.label-schema.schema-version="1.0.0-rc1" \
      maintainer="quentin.mcgaw@gmail.com" \
      org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/qdm12/REPONAME_GITHUB" \
      org.label-schema.url="https://github.com/qdm12/REPONAME_GITHUB" \
      org.label-schema.vcs-description="SHORT_DESCRIPTION" \
      org.label-schema.vcs-usage="https://github.com/qdm12/REPONAME_GITHUB/blob/master/README.md#setup" \
      org.label-schema.docker.cmd="docker run -d qmcgaw/REPONAME_DOCKER" \
      org.label-schema.docker.cmd.devel="docker run -it --rm qmcgaw/REPONAME_DOCKER" \
      org.label-schema.docker.params="" \
      org.label-schema.version="" \
      image-size="MB" \
      ram-usage="MB" \
      cpu-usage=""
# HEALTHCHECK --interval=300s --timeout=5s --start-period=5s --retries=1 CMD ["/healthcheck"]   
# USER 1000
ENTRYPOINT ["/dockerproxy"]
CMD -a containers

FROM builder AS builder2
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go test -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-s -w" -o app .
# RUN upx -v --best --ultra-brute --overlay=strip app && upx -t app

FROM final
COPY --from=builder2 /tmp/gobuild/app /dockerproxy