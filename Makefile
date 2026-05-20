build:
	go build ./cmd/gtd

install: test
	go install ./cmd/gtd

test:
	go test ./... -count=1

run: test
	go run ./cmd/gtd
