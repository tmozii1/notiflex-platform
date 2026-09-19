FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/notiflex-api .

FROM alpine:3.20

RUN addgroup -S notiflex && adduser -S notiflex -G notiflex
WORKDIR /app
COPY --from=builder /out/notiflex-api /app/notiflex-api

ENV PORT=8080
ENV APP_VERSION=v0.1.1

USER notiflex
EXPOSE 8080

ENTRYPOINT ["/app/notiflex-api"]
