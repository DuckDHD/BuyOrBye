# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.24.4-alpine AS build

ARG TARGETOS
ARG TARGETARCH

# ⬅️ Add libstdc++ and libgcc here
RUN apk add --no-cache curl ca-certificates libstdc++ libgcc && update-ca-certificates
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .
COPY ./configs ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/a-h/templ/cmd/templ@latest

RUN templ generate

# Choose correct Tailwind binary for musl + arch
RUN set -eux; \
    case "$TARGETARCH" in \
      amd64) TW_ARCH="x64-musl" ;; \
      arm64) TW_ARCH="arm64-musl" ;; \
      *) echo "Unsupported TARGETARCH: $TARGETARCH" >&2; exit 1 ;; \
    esac; \
    curl -fsSL "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-${TW_ARCH}" \
      -o /usr/local/bin/tailwindcss; \
    chmod +x /usr/local/bin/tailwindcss; \
    tailwindcss -i cmd/web/styles/input.css -o cmd/web/assets/css/output.css

ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o /app/main ./cmd/api/main.go

FROM alpine:3.20.1 AS prod
WORKDIR /app
COPY --from=build /app/main /app/main
# Copy the configs directory to the expected location
COPY --from=build /app/configs /app/configs
# COPY --from=build /app/cmd/web/assets /app/cmd/web/assets   # if you serve assets from disk
ENV PORT=8080
EXPOSE 8080
CMD ["/app/main"]