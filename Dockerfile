FROM golang:1.20-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o alilo-agent ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl

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
