run:
	go run main.go

tidy:
	go mod tidy

vendor:
	go mod vendor

build:
	goreleaser release --snapshot --skip-publish --clean --skip-sign