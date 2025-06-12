build-server:
	go build -o ./cmd/server/server ./cmd/server/

run-server:
	./cmd/server/server -d "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"

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

run-migrations:
	goose -dir ./migrations postgres "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable" up

run-docker-postgres:
	docker run --rm -d \
		--name yp-postgres-test \
		-e POSTGRES_USER=testuser \
		-e POSTGRES_PASSWORD=testpass \
		-e POSTGRES_DB=testdb \
		-p 5432:5432 \
		postgres:15-alpine
