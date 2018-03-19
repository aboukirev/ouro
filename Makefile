# Build Windows version
build: 
	go build ./cmd/oculeye
.PHONY: build

# Build Linux version
dist:
	GOOS=linux GOARCH=amd64 go build ./cmd/oculeye
.PHONY: dist
