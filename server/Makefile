dx-start:
	./scripts/start.sh

dx-stop:
	./scripts/stop.sh

dx-restart: dx-stop dx-start

aether: $(shell find . -name '*.go' -print)
	go build -o aether main.go

start: aether
	./aether
