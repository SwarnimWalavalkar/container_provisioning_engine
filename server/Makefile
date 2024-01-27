start:
	./scripts/start.sh

stop:
	./scripts/stop.sh

restart: stop start

aether: $(shell find . -name '*.go' -print)
	go build -o aether main.go

run: aether
	./aether 