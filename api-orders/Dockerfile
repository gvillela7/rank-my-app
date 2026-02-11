FROM golang:1.25.3-alpine AS builder
LABEL authors="gustavo"
ENV GOGC=75

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/main.go

FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/config.toml /app/config.toml
COPY --from=builder /app/docs /app/docs
EXPOSE 8000
CMD ["/app/api"]