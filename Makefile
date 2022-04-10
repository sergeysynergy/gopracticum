.PHONY: test

test:
	go test ./pkg/metrics/*
	go test ./internal/handlers/*
	go test ./internal/agent/*
