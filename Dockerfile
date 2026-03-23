FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o nucleus ./cmd/nucleus
RUN CGO_ENABLED=0 go build -o seed ./cmd/seed

FROM alpine:3.21
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY --from=builder /app/nucleus .
COPY --from=builder /app/seed .
COPY config.yaml .
COPY schemas/ schemas/

# Seed demo data on first run
RUN ./seed -repo data/repo -db data/nucleus.db

EXPOSE 8080
ENV NUCLEUS_BOOTSTRAP_SECRET=demo
CMD ["./nucleus", "--config", "config.yaml"]
