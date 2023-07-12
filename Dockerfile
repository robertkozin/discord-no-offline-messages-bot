# syntax=docker/dockerfile:1
FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

CMD ["go", "run", "main.go"]
