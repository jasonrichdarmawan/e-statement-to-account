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

# Copy required Go modules
COPY ./pdftotext ./pdftotext
COPY ./texttoparsed ./texttoparsed
COPY ./parsedtoaccount ./parsedtoaccount
COPY ./examples/api/main.go ./
RUN go build -o /main

FROM node:16.17-alpine3.16

WORKDIR /

# Compile .ts file to .js
COPY ./examples/api/app /app
RUN npm install typescript@4.8 -g
RUN npx tsc -p ./app --outDir ./public/js

FROM alpine:3.16

WORKDIR /

COPY --from=0 /main /main
COPY ./examples/api/public /public
COPY --from=1 /public/js /public/js

RUN apk add poppler-utils

RUN mkdir /secret-dir && chown 10001:10001 /secret-dir && chmod 750 /secret-dir

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