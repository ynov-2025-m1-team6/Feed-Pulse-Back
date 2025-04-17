FROM golang:1.24-bookworm AS buildergo
LABEL authors="tot0p"


ENV PATH=/usr/local/bin:$PATH

ENV LANG=C.UTF-8

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o /app/app ./cmd/app/main.go

FROM debian:bookworm-slim
LABEL authors="tot0p"


ENV LANG=C.UTF-8

WORKDIR /app

COPY --from=buildergo /app/app .

ENV PORT=8080
EXPOSE 8080



CMD ["/app/app"]
