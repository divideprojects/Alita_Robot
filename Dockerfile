# Build Stage: Build bot using the alpine image, also install doppler in it
FROM golang:1.20-alpine AS builder
RUN apk add --no-cache curl wget gnupg git
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=`go env GOHOSTOS` GOARCH=`go env GOHOSTARCH` go build -o out/Alita_Robot -ldflags="-w -s" .

# Run Stage: Run bot using the bot binary copied from build stage
FROM alpine:3.18.0
COPY --from=builder /app/out/Alita_Robot /app/Alita_Robot
CMD ["/app/Alita_Robot"]
