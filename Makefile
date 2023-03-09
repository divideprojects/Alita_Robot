run:
	go run main.go

tidy:
	go mod tidy

vendor:
	go mod vendor

docker-compose:
	set -x DOPPLER_TOKEN (doppler configs tokens create dev --plain --max-age 1m) && docker compose down && docker compose up -d --build
