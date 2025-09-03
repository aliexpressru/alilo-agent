FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o alilo-agent ./cmd/main.go

FROM ghcr.io/grafana/k6:1.2.3

ENTRYPOINT []

USER root
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/alilo-agent .
COPY --from=builder /app/config.json .
COPY --from=builder /app/resources ./resources

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8888

CMD ["./alilo-agent", "-serverPort=8888"]
