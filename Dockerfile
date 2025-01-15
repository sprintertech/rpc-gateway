FROM golang:1.22-alpine3.19 AS builder

RUN apk add --update-cache \
        git \
        build-base

WORKDIR /src
COPY . .

RUN go build .

FROM alpine:3.19

RUN apk add --update-cache --no-cache \
        ca-certificates

COPY --from=builder /src/rpc-gateway /app/

USER nobody
LABEL org.opencontainers.image.source https://github.com/sprintertech/rpc-gateway
ENTRYPOINT ["/app/rpc-gateway"]