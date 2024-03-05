dx-start:
	./scripts/start.sh

dx-stop:
	./scripts/stop.sh

dx-restart: dx-stop dx-start

container_provisioning_engine: $(shell find . -name '*.go' -print)
	go build -o container_provisioning_engine main.go

start: container_provisioning_engine
	./container_provisioning_engine
