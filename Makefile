default: all

all: tidy lint test

clean:
	rm -fr .coverage.out

lint:
	go vet ./...
	golangci-lint run

test:
	gotest -coverprofile .coverage.out -race -timeout 10s

tidy:
	go mod tidy
