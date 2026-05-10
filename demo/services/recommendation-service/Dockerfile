FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o recommendation-service ./demo/services/recommendation-service/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/recommendation-service ./recommendation-service
COPY --from=builder /app/demo/services/recommendation-service/configs ./configs

EXPOSE 9103
ENV CONFIGS_DIR=./configs
CMD ["./recommendation-service"]
