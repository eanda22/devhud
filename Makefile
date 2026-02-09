.PHONY: build run test lint fmt clean

build:
	go build -o devhud .

run:
	go run .

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofmt -w .

clean:
	rm -f devhud