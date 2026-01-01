# Gen-SQL

## Guide

### Dependencies (both BE and FE)
```bash
make deps
```

### BE
```bash
make up
```

### FE
```bash
cd web/
npm run dev
```

## PROTO
```bash
make proto-tools # for initial use
make proto-gen
```

## MIGRATIONS

Manual :c  ./temp/schema.sql

## DOCKER IMAGE

```bash
docker pull ghcr.io/kacperborowieckb/gen-sql/api:latest
```

## MONITORING

### Metrics Endpoint

The API service exposes metrics at: `http://localhost:8080/metrics`

### Available Metrics

- `http_requests_total` - Total number of HTTP requests (with labels: method, route, status_code)
- `http_request_duration_seconds` - HTTP request duration histogram
- `http_requests_in_flight` - Number of requests currently being processed
