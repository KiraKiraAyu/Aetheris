# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies for compiling go-sqlite3 (CGO)
RUN apk add --no-cache build-base

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build API and Worker binaries with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/api ./cmd/api
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/worker ./cmd/worker

# Runtime stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy compiled binaries from builder
COPY --from=builder /app/api /app/api
COPY --from=builder /app/worker /app/worker

# Expose API port
EXPOSE 8080

# Default command
CMD ["/app/api"]
