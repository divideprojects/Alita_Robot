# Build Stage: Build bot using the alpine image, also install doppler in it
FROM golang:1.20-alpine AS builder
RUN apk add --no-cache curl wget gnupg git
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=`go env GOHOSTOS` GOARCH=`go env GOHOSTARCH` go build -o out/Alita_Robot -ldflags="-w -s" .

# Run Stage: Run bot using the bot binary copied from build stage
FROM alpine:3.17.2
COPY --from=builder /app/out/Alita_Robot /app/Alita_Robot
RUN wget -q -t3 'https://packages.doppler.com/public/cli/rsa.8004D9FF50437357.key' -O /etc/apk/keys/cli@doppler-8004D9FF50437357.rsa.pub \
    && echo 'https://packages.doppler.com/public/cli/alpine/any-version/main' | tee -a /etc/apk/repositories \
    && apk add --no-cache doppler
CMD ["doppler", "run", "--", "/app/Alita_Robot"]
