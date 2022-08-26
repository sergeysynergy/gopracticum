.PHONY: test

test:
	go test ./pkg/exitcheck/*
	go test ./pkg/metrics/*
	go test ./internal/agent/*
	go test ./internal/data/repository/pgsql/*
	go test ./internal/filestore/*
	go test ./internal/handlers/*
	go test ./internal/httpserver/*
	go test ./internal/storage/*

cover:
	go test -cover ./pkg/exitcheck/*
	go test -cover ./pkg/metrics/*
	go test -cover ./internal/agent/*
	go test -cover ./internal/data/repository/pgsql/*
	go test -cover ./internal/filestore/*
	go test -cover ./internal/handlers/*
	go test -cover ./internal/httpserver/*
	go test -cover ./internal/storage/*
