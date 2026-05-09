FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./demo/services/user-service/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/user-service ./user-service
COPY --from=builder /app/demo/services/user-service/configs ./configs

EXPOSE 9101
ENV CONFIGS_DIR=./configs
CMD ["./user-service"]
