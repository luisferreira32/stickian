FROM golang:1.25-alpine AS builder
WORKDIR /root

# install dependencies and run the server for development
COPY go.mod go.sum .
RUN go mod download
COPY server/ ./server/
EXPOSE 8080
ENV CGO_ENABLED=0
CMD ["go", "run", "./server/"]
