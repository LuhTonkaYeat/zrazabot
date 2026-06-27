BINARY_NAME=zrazabot
GO=go

build:
	GO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY_NAME) .
	git add *.go

run:
	$(GO) run main.go

clean:
	rm -f $(BINARY_NAME)

.PHONY: build run clean