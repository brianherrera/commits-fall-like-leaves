.PHONY: build clean deploy security

security:
	go vet ./cmd/... ./internal/...
	gosec ./cmd/... ./internal/...

build: security
	cd cmd/haiku && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap main.go
	zip -j lambda-function.zip cmd/haiku/bootstrap
	rm cmd/haiku/bootstrap

clean:
	rm -f lambda-function.zip
	rm -f cmd/haiku/bootstrap
