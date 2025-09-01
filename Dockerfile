# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata file
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

# Go deps
COPY go.mod go.sum ./
RUN go mod download

# Source
COPY . .

# Generate templ files
RUN templ generate

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/app ./cmd/server
RUN ls -la bin/app && file bin/app

# Final stage
FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -g 1001 -S app && adduser -u 1001 -S app -G app

# Put binary in PATH (not under /app to avoid bind-mount collisions)
COPY --from=builder /app/bin/app /usr/local/bin/app
RUN chmod +x /usr/local/bin/app

# App working dir holds only data/assets
WORKDIR /srv
COPY --from=builder /app/static /srv/static
COPY --from=builder /app/migrations /srv/migrations

RUN chown -R app:app /srv
USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Use PATH so it's immune to /srv mounts
CMD ["app"]
