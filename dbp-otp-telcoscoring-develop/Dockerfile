FROM golang:1.18-alpine as builder
WORKDIR /build
COPY migrations .
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main cmd/app/main.go
FROM alpine:3
WORKDIR /app
COPY migrations /app/migrations 
COPY --from=builder main /app/main
ENTRYPOINT ["/app/main"]

