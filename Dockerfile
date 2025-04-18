FROM golang:1.24-bookworm AS buildergo
LABEL authors="tot0p"


ENV PATH=/usr/local/bin:$PATH

ENV LANG=C.UTF-8

WORKDIR /app

COPY . .

RUN go mod download

RUN go run github.com/swaggo/swag/cmd/swag@latest init -g main.go --output docs/feed-pulse --dir ./cmd/app,./internal/api/handlers,./internal/utils

RUN go build -o /app/app ./cmd/app/main.go

FROM debian:bookworm-slim
LABEL authors="tot0p"


ENV LANG=C.UTF-8

WORKDIR /app

COPY --from=buildergo /app/app .
COPY --from=buildergo /app/docs/feed-pulse/swagger.json ./docs/feed-pulse/swagger.json
COPY --from=buildergo /app/docs/feed-pulse/swagger.yaml ./docs/feed-pulse/swagger.yaml

ENV PORT=8080
EXPOSE 8080

CMD ["/app/app"]
