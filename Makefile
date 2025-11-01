SHELL := /bin/sh

.PHONY: deps tidy build up stop down logs api data generator query proto-tools proto-gen

deps:
	go get github.com/go-chi/chi/v5@v5.0.11
	go get github.com/rabbitmq/amqp091-go@v1.10.0
	go get github.com/go-playground/validator/v10@v10.20.0
	go get github.com/lib/pq@v1.10.9
	go get google.golang.org/grpc@latest
	go get google.golang.org/protobuf@latest
	go mod tidy

.PHONY: proto-tools
proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: proto-gen
proto-gen:
	@mkdir -p shared/gen
	@echo "Generating protobuf Go code..."
	protoc --go_out=./shared/gen --go_opt=paths=source_relative \
	       --go-grpc_out=./shared/gen --go-grpc_opt=paths=source_relative \
	       proto/*.proto

build: proto-gen
	docker build -f services/api/Dockerfile -t gensql-api .
	docker build -f services/data/Dockerfile -t gensql-data .
	docker build -f services/generator/Dockerfile -t gensql-generator .
	docker build -f services/query/Dockerfile -t gensql-query .

up:
	docker compose up --build -d

stop:
	docker compose stop

down:
	docker compose down -v

logs:
	docker compose logs -f --tail=200

api:
	PORT=8080 go run ./services/api

data:
	PORT=8081 go run ./services/data

generator:
	PORT=8082 go run ./services/generator

query:
	PORT=8083 go run ./services/query
