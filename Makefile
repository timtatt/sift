.PHONY: test clean

vet:
	go vet ./cmd/... ./pkg/... ./internal/...

test:
	go test ./cmd/... ./pkg/... ./internal/...

