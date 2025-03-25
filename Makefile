.PHONY:
run-spidey:
	go run ./cmd/spidey/main.go $(args)

.PHONY:
build-spidey:
	go build -o spidey ./cmd/spidey/main.go 
