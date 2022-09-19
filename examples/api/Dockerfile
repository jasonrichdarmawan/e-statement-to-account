# syntax=docker/dockerfile:1

## Build
# Alpine is chosen for its small footprint
# compared to Ubuntu
FROM golang:1.19-alpine AS stage1

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

FROM node:16.17-alpine3.16 AS stage2

WORKDIR /

# Compile .ts file to .js
COPY ./examples/api/app /app
RUN npm install typescript@4.8 -g
RUN npx tsc -p ./app --outDir ./public/js

FROM alpine:3.16 AS stage3

WORKDIR /

RUN apk add poppler-utils

RUN adduser \
    -h "/dev/null" \
    -g "" \
    -s "/sbin/nologin" \
    -D \
    -H \
    -u 10001 \
    playerone

RUN mkdir /secret-dir && chown 10001:10001 /secret-dir && chmod 750 /secret-dir

FROM scratch AS stage4

WORKDIR /

# aarch64
ARG ARCH

# $ ldd pdftotext
COPY --from=stage3 /lib/ld-musl-${ARCH}.so.1 \
/lib/ld-musl-${ARCH}.so.1 \
/lib/libz.so.1 \
/lib/
COPY --from=stage3 /usr/lib/libpoppler.so.121 \
/usr/lib/libstdc++.so.6 \
usr/lib/libfreetype.so.6 \
/usr/lib/libfontconfig.so.1 \
/usr/lib/libjpeg.so.8 \
/usr/lib/libopenjp2.so.7 \
/usr/lib/liblcms2.so.2 \
/usr/lib/libpng16.so.16 \
/usr/lib/libtiff.so.5 \
usr/lib/libgcc_s.so.1 \
/usr/lib/libbz2.so.1 \
/usr/lib/libbrotlidec.so.1 \
/usr/lib/libexpat.so.1 \
usr/lib/libwebp.so.7 \
/usr/lib/libzstd.so.1 \
/usr/lib/liblzma.so.5 \
/usr/lib/libbrotlicommon.so.1 \
/usr/lib/
COPY --from=stage3 /usr/bin/pdftotext /usr/bin/pdftotext

COPY --from=stage1 /main /main

COPY ./examples/api/public/ /public/
COPY --from=stage2 /public/js/ /public/js/

COPY --from=stage3 /secret-dir/ /secret-dir/
COPY --from=stage3 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=stage3 --chown=10001:10001 /tmp/ /tmp/

COPY --from=stage3 /etc/passwd /etc/passwd
# USER 10001 works. However, the gid=0(root). So, don't do USER 10001.
USER 10001:10001

EXPOSE 80/tcp 443/tcp