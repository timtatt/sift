.PHONY: test clean

clean:
	rm -rf build/

test:
	go test ./cmd/... ./pkg/... ./internal/...

build: clean
	env GOOS=linux GOARCH=amd64 go build -o build/sift_linux_amd64 main.go
	env GOOS=linux GOARCH=arm64 go build -o build/sift_linux_arm64 main.go
	env GOOS=darwin GOARCH=arm64 go build -o build/sift_darwin_arm64 main.go
	env GOOS=darwin GOARCH=amd64 go build -o build/sift_darwin_amd64 main.go
