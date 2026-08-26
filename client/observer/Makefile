.DEFAULT_GOAL := build

.PHONY:fmt vet build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

clean: 
	go clean ./...
	
build: vet
	go build