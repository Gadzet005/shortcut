FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mock-service ./tests/mock-service/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/mock-service ./mock-service
COPY --from=builder /app/tests/mock-service/configs ./tests/mock-service/configs

EXPOSE 9001
ENV CONFIGS_DIR=./tests/mock-service/configs
CMD ["./mock-service"]
