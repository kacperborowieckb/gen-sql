SHELL := /bin/sh

.PHONY: deps tidy build up stop down logs api data generator query

deps:
	go get github.com/go-chi/chi/v5@v5.0.11
	go get github.com/rabbitmq/amqp091-go@v1.10.0
	go get github.com/go-playground/validator/v10@v10.20.0
	go get github.com/lib/pq@v1.10.9 
	go mod tidy

build:
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

