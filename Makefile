.PHONY: run build clean

BINARY := ./bin/snakedex
CMD    := ./cmd/server

run:
	go run $(CMD)

build:
	go build -o $(BINARY) $(CMD)

clean:
	rm -f $(BINARY)
