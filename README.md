# Gen-SQL

## start
```bash
make deps
docker compose up --build -d
# check health (only 8080 and 8083 are exposed, 8081 and 8082 accept internal traffic)
curl -s localhost:8080/healthz
curl -s localhost:8083/healthz
```

## Environment
- DATABASE_URL: `postgres://postgres:postgres@db:5432/gensql?sslmode=disable`
- AMQP_URL: `amqp://guest:guest@rabbitmq:5672/`
- GENERATOR_QUEUE: `gensql.jobs`

## single service run
```bash
make api
make data
make generator
make query
```