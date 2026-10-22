FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/app /app/app
COPY --from=builder /app/web /app/web

EXPOSE 8080

CMD ["/app/app"]
