# Docker-proxy-ACL-Alpine

Lightweight container running a restricted Docker unix socket proxy

This is heavily inspired from [titpetric/docker-proxy-acl](https://github.com/titpetric/docker-proxy-acl) which runs it with a 480MB Docker image

This Docker image is only **6MB** and settings can be adjusted with the environment variable `OPTIONS`

[![Build Status](https://travis-ci.org/qdm12/docker-proxy-acl-alpine.svg?branch=master)](https://travis-ci.org/qdm12/docker-proxy-acl-alpine)
[![Docker Build Status](https://img.shields.io/docker/build/qmcgaw/docker-proxy-acl-alpine.svg)](https://hub.docker.com/r/qmcgaw/docker-proxy-acl-alpine)

[![GitHub last commit](https://img.shields.io/github/last-commit/qdm12/docker-proxy-acl-alpine.svg)](https://github.com/qdm12/docker-proxy-acl-alpine/commits)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/y/qdm12/docker-proxy-acl-alpine.svg)](https://github.com/qdm12/docker-proxy-acl-alpine/commits)
[![GitHub issues](https://img.shields.io/github/issues/qdm12/docker-proxy-acl-alpine.svg)](https://github.com/qdm12/docker-proxy-acl-alpine/issues)

[![Docker Pulls](https://img.shields.io/docker/pulls/qmcgaw/docker-proxy-acl-alpine.svg)](https://hub.docker.com/r/qmcgaw/docker-proxy-acl-alpine)
[![Docker Stars](https://img.shields.io/docker/stars/qmcgaw/docker-proxy-acl-alpine.svg)](https://hub.docker.com/r/qmcgaw/docker-proxy-acl-alpine)
[![Docker Automated](https://img.shields.io/docker/automated/qmcgaw/docker-proxy-acl-alpine.svg)](https://hub.docker.com/r/qmcgaw/docker-proxy-acl-alpine)

[![](https://images.microbadger.com/badges/image/qmcgaw/docker-proxy-acl-alpine.svg)](https://microbadger.com/images/qmcgaw/docker-proxy-acl-alpine)
[![](https://images.microbadger.com/badges/version/qmcgaw/docker-proxy-acl-alpine.svg)](https://microbadger.com/images/qmcgaw/docker-proxy-acl-alpine)

| Download size | Image size | RAM usage | CPU usage |
| --- | --- | --- | --- |
| 5MB | 6MB | 10MB | Low |



## Why

Exposing `/var/run/docker.sock` to a Docker container requiring it (such as [netdata](https://github.com/firehol/netdata)) involves
security concerns and the container should be limited in what it can do with `docker.sock`.

You can enable an endpoint with the `-a` argument. Currently supported endpoints are:

- containers: opens access to `/containers/json` and `/containers/{name}/json`
- images: opens access to `/images/json` , `/images/{name}/json` and `/images/{name}/history`
- networks: opens access to `/networks` and `/networks/{name}`
- volumes: opens access to `/volumes` and `/volumes/{name}`
- services: opens access to `/services` and `/services/{id}`
- tasks: opens access to `/tasks` and `/tasks/{name}`
- events: opens access to `/events`
- info: opens access to `/info`
- version: opens access to `/version`
- ping: opens access to `/_ping`

To combine arguments, repeat them like this: `./run -a info -a version`.

## Example usage: limiting access for the netdata container

The project [**netdata**](https://github.com/firehol/netdata) can use the `docker.sock` file to resolve
the container names found in the `cgroups` filesystem, into readable names.

In this case, run this container with:

```bash
docker run -d -e OPTIONS="-a containers" -v /var/run/docker.sock:/var/run/docker.sock \
-v /yourpath:/tmp/docker-proxy-acl --net=none qmcgaw/docker-proxy-acl-alpine
```



A new socket file is hence created at `/yourpath/docker.sock` with only the
`/containers/json` and `/containers/{name}/json` endpoints allowed.


This socket file can then be passed to the **netdata** container, with an additional option like this:

```bash
-v /yourpath/docker.sock:/var/run/docker.sock
```



You can also use docker-compose:


```yml
version: '3'
services:
  docker-proxy:
    build: .
    image: qmcgaw/docker-proxy-acl-alpine
    container_name: docker-proxy
    volumes:
      - /tmp/docker-proxy-acl:/tmp/docker-proxy-acl
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - OPTIONS=-a containers
    network_mode: none
    restart: always
```
