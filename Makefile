build:
	CGO_ENABLED=0 go build -o headcount ./cmd/headcount/

run: build
	./headcount

test:
	go test ./...

clean:
	rm -f headcount

.PHONY: build run test clean
