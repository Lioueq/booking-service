FROM golang:1.26.1 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

FROM debian:stable-slim
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/* && \
    useradd -m -u 1000 appuser

WORKDIR /app
COPY --from=builder --chown=appuser:appuser /app/server /app/
COPY --from=builder --chown=appuser:appuser /app/configs /app/configs
COPY --from=builder --chown=appuser:appuser /app/docs /app/docs

USER appuser
EXPOSE 8080

CMD ["/app/server"]