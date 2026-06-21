.PHONY: build run test fmt vet tidy

build:
	go build -o termipost .

run:
	go run .

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy
