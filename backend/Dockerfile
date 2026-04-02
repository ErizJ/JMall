# Build stage
FROM golang:1.23-alpine AS builder

ARG SERVICE
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./service/${SERVICE}/${SERVICE}.go

# Runtime stage
FROM alpine:3.20

ARG SERVICE
ENV SERVICE=${SERVICE}

RUN apk add --no-cache tzdata ca-certificates
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /service /app/service
COPY service/${SERVICE}/etc /app/etc

# Substitute environment variables into config at startup
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

ENTRYPOINT ["/app/docker-entrypoint.sh"]
