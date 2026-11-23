# syntax=docker/dockerfile:1
#
# Multi-stage build producing a small, static, non-root image. Base images are
# pinned by digest for reproducible builds; Dependabot keeps the digests current.

FROM golang:1.27.4-trixie@sha256:b54b2639de8c1bfe538ec5ffc45f29b1932235ac5f6d63c059853646677112d9 AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
ARG VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags "-s -w -X github.com/y0f/gophling/v2/internal/config.version=${VERSION}" -o /out/api ./cmd/api && \
    go build -trimpath -ldflags "-s -w -X github.com/y0f/gophling/v2/internal/config.version=${VERSION}" -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian13:nonroot@sha256:8ce6dd87e9823c5b48dcfb2c8604714ccb351ce59607327bcc4957448da279b5
COPY --from=build /out/api /api
COPY --from=build /out/migrate /migrate
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
