FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o product-service ./demo/services/product-service/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/product-service ./product-service
COPY --from=builder /app/demo/services/product-service/configs ./configs

EXPOSE 9102
ENV CONFIGS_DIR=./configs
CMD ["./product-service"]
