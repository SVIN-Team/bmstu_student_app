FROM golang:1.25-alpine AS builder

WORKDIR /src

# Install build tools
RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/studhub ./cmd/main.go

# Runtime image
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/studhub /app/studhub
COPY deploy /app/deploy

ENV PATH="/app:${PATH}"

ENTRYPOINT ["/app/studhub"]
CMD ["--config", "/app/deploy/config.yaml"]
