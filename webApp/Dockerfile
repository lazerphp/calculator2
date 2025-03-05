# syntax=docker/dockerfile:1

FROM golang:latest AS builder

WORKDIR /app

COPY . ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/static ./static

EXPOSE 80

CMD ["/app/app"]
