build-server:
	go build -o ./cmd/server/server ./cmd/server/

run-server:
	./cmd/server/server

build-agent:
	go build -o ./cmd/agent/agent ./cmd/agent/

run-agent:
	./cmd/agent/agent

mockgen:	
	mockgen -source=$(file) \
		-destination=$(dir $(file))$(notdir $(basename $(file)))_mock.go \
		-package=$(shell basename $(dir $(file)))

test:
	go test -cover ./... 
