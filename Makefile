run:
	@go run cmd/server/main.go
build.docker:
	@docker build -f build/pkg/Dockerfile -t kod-server .
