FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o shortcut ./cmd/shortcut/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/shortcut ./shortcut
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/tests/configs ./tests/configs

EXPOSE 8080
ENTRYPOINT ["./shortcut"]
CMD ["-graphconfigs", "./tests/configs"]
