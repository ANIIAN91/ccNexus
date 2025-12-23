# syntax=docker/dockerfile:1

FROM golang:1.24.3-bookworm AS build

WORKDIR /src

# Network-reliability knobs (override at build time if needed)
# Example:
#   docker build --build-arg GOPROXY=https://goproxy.cn,direct --build-arg GOSUMDB=sum.golang.google.cn ...
ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn
ENV GOPROXY=$GOPROXY
ENV GOSUMDB=$GOSUMDB

# Cache dependencies
# NOTE: repo may not have go.sum; it will be generated during download.
COPY go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy source
COPY . .

# Ensure go.sum is complete (repo may not include it)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Build the headless API server (webui is embedded via go:embed)
# modernc.org/sqlite is pure Go, so we can build a static binary without gcc.
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -mod=mod -trimpath -ldflags "-s -w" -o /out/ccnexus-server ./cmd/server

# Runtime stage
# Note: using :nonroot can fail to create/write files in a fresh named volume mounted at /data.
# Run as root by default for better out-of-the-box usability.
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=build /out/ccnexus-server /app/ccnexus-server

ENV CCNEXUS_DATA_DIR=/data
ENV CCNEXUS_PORT=3000
ENV CCNEXUS_DB_PATH=/data/ccnexus.db

VOLUME ["/data"]
EXPOSE 3000

ENTRYPOINT ["/app/ccnexus-server"]
