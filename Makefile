.PHONY: run tidy vendor build

GO_CMD = go
GORELEASER_CMD = goreleaser

run:
    $(GO_CMD) run main.go

tidy:
    $(GO_CMD) mod tidy

vendor:
    $(GO_CMD) mod vendor

build:
    $(GORELEASER_CMD) release --snapshot --skip=publish --clean --skip=sign
