.PHONY: test

test:
	go test ./...

cover:
	go test -cover ./...
