FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o english_bot .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/english_bot ./english_bot

CMD ["./english_bot"]