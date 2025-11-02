FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod ./
COPY shared ./shared
COPY utils ./utils
COPY services/data ./services/data

# cache mod and go-build to not re-download all dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod tidy && cd services/data && go build -ldflags="-s -w" -o /out/app ./

# ------

FROM gcr.io/distroless/base:nonroot

USER 65532:65532

COPY --from=builder /out/app /app

CMD ["/app"]

