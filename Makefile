.PHONY: test

test:
	go test -v ./pkg/metrics/*
	go test -v ./internal/handlers/*
	go test -v ./internal/agent/*
