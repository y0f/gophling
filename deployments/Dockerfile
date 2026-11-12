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

FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329
COPY --from=build /out/api /api
COPY --from=build /out/migrate /migrate
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
