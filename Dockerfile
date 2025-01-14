ARG RAW_VERSION=v0.1.0

FROM golang:1.23-alpine AS builder

ARG RAW_VERSION

RUN apk add --no-cache gcc musl-dev git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build -ldflags "-X main.RawVersion='${RAW_VERSION}'" -o ants github.com/probe-lab/ants-watch/cmd/ants

FROM alpine:3.18

RUN adduser -D -H ants-watch

WORKDIR /home/ants-watch
USER ants-watch

COPY --from=builder /build/ants /usr/local/bin/ants

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s CMD ants health

CMD ["ants"]
