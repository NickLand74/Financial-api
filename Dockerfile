FROM golang:1.21-alpine

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o financial-api ./cmd/server/main.go

CMD ["./financial-api"]