FROM node:22-alpine AS frontend-builder
WORKDIR /app
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

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
COPY --from=frontend-builder /app/dist ./web/dist

EXPOSE 8080
ENTRYPOINT ["./shortcut"]
CMD ["-graphconfigs", "./tests/configs"]
