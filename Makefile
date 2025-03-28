.PHONY:
build:
	go build -o nomad ./cmd/nomad/main.go 

.PHONY:
run:
	go run ./cmd/nomad/main.go $(args)

