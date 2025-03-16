FROM golang:1.24.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine

WORKDIR /app

RUN adduser -D appuser

USER appuser

COPY --from=builder /app/config ./config

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
