FROM golang:1.25-alpine AS builder
WORKDIR /app

# install dependencies and build the server
COPY go.mod ./
RUN go mod download
COPY server/ ./server/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o stickian-server ./server

# deploy without golang
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/stickian-server .
EXPOSE 8080
CMD ["./stickian-server"]
