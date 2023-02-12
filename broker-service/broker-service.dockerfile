FROM golang:1.18-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o brokerservice ./cmd/api

RUN chmod +x /app/brokerservice

#Build a tiny docker image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/brokerservice /app

CMD ["/app/brokerservice"]