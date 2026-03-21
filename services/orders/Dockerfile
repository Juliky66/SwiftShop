FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum* ./
COPY shared/ ./shared/
COPY services/auth/ ./services/auth/
COPY services/catalog/ ./services/catalog/
COPY services/orders/ ./services/orders/
COPY services/seller/ ./services/seller/
COPY services/reviews/ ./services/reviews/

WORKDIR /app/services/orders
RUN go build -o /orders ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /orders .
EXPOSE 8083
ENTRYPOINT ["/app/orders"]
