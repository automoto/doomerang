lint:
	golangci-lint run

run:
	go run main.go

build:
	go build .

basic-test:
	./scripts/basic-test.sh