.PHONY: test-unit test test-integration test-all

test-unit:
	@echo "Running default mock/local tests..."
	go test ./... -v -race -cover

test: test-unit

test-integration:
	@echo "Running real API smoke tests..."
	go test ./... -tags=integration -run Integration -v -count=1

test-all: test test-integration
