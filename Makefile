build-server:
	go build -o ./cmd/server/server ./cmd/server/

run-server:
	./cmd/server/server

mockgen:	
	mockgen -source=$(file) \
		-destination=$(dir $(file))$(notdir $(basename $(file)))_mock.go \
		-package=$(shell basename $(dir $(file)))

test:
	go test -cover ./... 
