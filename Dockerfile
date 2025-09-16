# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.24.4-alpine AS build

ARG TARGETOS
ARG TARGETARCH

# Node.js + npm for Tailwind build
RUN apk add --no-cache curl ca-certificates libstdc++ libgcc nodejs npm && update-ca-certificates

WORKDIR /app

# ---------------- Go deps ----------------
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# ---------------- App source ----------------
COPY . .
COPY ./configs ./

# ---------------- templ ----------------
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate

# ---------------- Frontend (Tailwind) ----------------
# Install deps declared in cmd/web/package.json (plugins like @tailwindcss/forms resolve here)
RUN --mount=type=cache,target=/root/.npm npm ci --prefix cmd/web

# Optional: quiet Browserslist warning in CI (safe to skip)
# RUN npx --yes --prefix cmd/web update-browserslist-db@latest

# Build CSS using local tailwind + config
RUN npx --prefix cmd/web tailwindcss \
  -c cmd/web/tailwind.config.js \
  -i cmd/web/src/css/input.css \
  -o cmd/web/assets/css/output.css

# ---------------- Backend build ----------------
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o /app/main ./cmd/app/main.go


# ================= Production image =================
FROM alpine:3.20.1 AS prod
WORKDIR /app

# Binary
COPY --from=build /app/main /app/main

# Configs
COPY --from=build /app/configs /app/configs

# Frontend assets needed at runtime
# - assets: built CSS/JS (your server can reference them directly)
# - static: served by router.Static("/static", "cmd/web/static")
COPY --from=build /app/cmd/web/assets  /app/cmd/web/assets
COPY --from=build /app/cmd/web/static  /app/cmd/web/static

ENV PORT=8080
EXPOSE 8080
CMD ["/app/main"]
