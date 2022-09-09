# syntax=docker/dockerfile:1

# Alpine is chosen for its small footprint
# compared to Ubuntu
FROM golang:1.19-alpine AS build

WORKDIR /app

# Download necessary Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /main examples/api/main.go

FROM alpine:3.16

RUN apk add poppler-utils

WORKDIR /

COPY --from=build /main /main
COPY --from=build /app/examples/api/public/ /public/

RUN adduser \
--disabled-password \
--gecos "" \
--home "/nonexistent" \
--shell "/sbin/nologin" \
--no-create-home \
--uid "10001" \
"appuser"
# Use an underprivileged user.
USER appuser:appuser

EXPOSE 80/tcp 443/tcp

ENTRYPOINT ["/main"]