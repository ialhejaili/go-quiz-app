FROM golang:1.22.3 AS builder
WORKDIR /app
COPY go.mod go.sum .env ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o go-quiz-app


FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/go-quiz-app ./
COPY --from=builder /app/.env ./
CMD ["./go-quiz-app"]
