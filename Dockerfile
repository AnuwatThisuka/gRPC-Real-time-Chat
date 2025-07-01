FROM golang:1.21

WORKDIR /app
COPY . .

RUN go mod tidy && make generate

CMD ["go", "run", "server/main.go"]
