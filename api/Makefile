generate-mock:
	mockgen -source=internal/invoice/repository.go -destination=internal/invoice/repository_mock.go -package=invoice

lint:
	golangci-lint run ./...

test:
	go test ./...
