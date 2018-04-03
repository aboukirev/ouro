# Build Windows version
build: 
	go build ./cmd/ouro
.PHONY: build

# Build Linux version
dist:
	GOOS=linux GOARCH=amd64 go build ./cmd/ouro
.PHONY: dist
