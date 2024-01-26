start:
	./scripts/start.sh

stop:
	./scripts/stop.sh

restart: stop start

aether: $(wildcard *.go **/*.go)
	go build -o aether main.go


run: aether
	./aether 