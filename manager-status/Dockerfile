FROM golang:1.25.3-alpine AS builder
LABEL authors="gustavo"
ENV GOGC=75

WORKDIR /manager
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /manager/bin/appmanager ./cmd/main.go

FROM gcr.io/distroless/static-debian12

WORKDIR /manager
COPY --from=builder /manager/bin/appmanager /manager/appmanager
COPY --from=builder /manager/config.toml /manager/config.toml

CMD ["/manager/appmanager"]