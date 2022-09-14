# syntax=docker/dockerfile:1

## Build
# Alpine is chosen for its small footprint
# compared to Ubuntu
FROM golang:1.19-alpine

WORKDIR /app

# Download necessary Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /main ./examples/api/main.go

FROM alpine:3.16

WORKDIR /

COPY --from=0 /main /main

COPY --from=0 /app/examples/api/public /public

RUN apk add poppler-utils

RUN mkdir /secret-dir && chown 10001:10001 /secret-dir & chmod 750 /secret-dir

RUN adduser \
    -h "/dev/null" \
    -g "" \
    -s "/sbin/nologin" \
    -D \
    -H \
    -u 10001 \
    playerone
# USER 10001 works. However, the gid=0(root). So, don't do USER 10001.
USER 10001:10001

EXPOSE 80/tcp 443/tcp

ENTRYPOINT ["/main", "-env", "production", "-hostname", "estatement.godata.id"]