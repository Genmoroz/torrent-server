include ./linting.mk

.PHONY: deps
deps:
	go mod tidy -v
	go mod download
	go mod vendor -v
	go mod verify
