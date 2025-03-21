.PHONY:
run-spidey:
	go run ./cmd/spidey/main.go

.PHONY:
spidey:
	go build -o spidey ./cmd/spidey/main.go 
