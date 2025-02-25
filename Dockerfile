# Build the landing page
FROM oven/bun:1 as landing-builder
WORKDIR /app/web/landing-page
COPY web/landing-page/package.json web/landing-page/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/landing-page .
RUN bun run build

# Build the Go application
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=landing-builder /app/web/landing-page/dist ./web/landing-page/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /gitignore-lol ./cmd/main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates git
WORKDIR /root/
COPY --from=builder /gitignore-lol ./
EXPOSE 4444 
ENTRYPOINT ["./gitignore-lol"] 