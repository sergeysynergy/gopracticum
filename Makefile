.PHONY: test

test:
	go test ./pkg/metrics/*
	go test ./internal/filestore/*
	go test ./internal/handlers/*
	go test ./internal/agent/*
