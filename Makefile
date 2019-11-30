
build:
	go build cmd/mute.go


install:
	go install cmd/mute.go


test:
	go test github.com/farzadghanei/mute


.DEFAULT_GOAL := build
.PHONY: test build install

