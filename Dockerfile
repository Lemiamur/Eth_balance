FROM golang:1.22.4 AS builder
WORKDIR /app
COPY . .
RUN go build -o eth_bal ./cmd/app/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/eth_bal .
COPY .env .
CMD ["./eth_bal"]
