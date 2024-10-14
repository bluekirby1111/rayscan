FROM golang:1.23-alpine AS builder
WORKDIR /usr/local/rayscan/
COPY . .
RUN go mod download
RUN go build -o rayscan -ldflags "-s -w" main.go

FROM debian:bookworm-slim AS binaries
WORKDIR /usr/local/rayscan/
RUN DEBIAN_FRONTEND=noninteractive apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get \
    -o Dpkg::Options::=--force-confold \
    -o Dpkg::Options::=--force-confdef \
    -y -q --allow-downgrades --allow-remove-essential --allow-change-held-packages \
    install ca-certificates
COPY --from=builder /usr/local/rayscan/rayscan /usr/local/bin/
COPY --from=builder /usr/local/rayscan/config.toml /usr/local/rayscan/config.toml

EXPOSE 80

CMD [ "rayscan" ]