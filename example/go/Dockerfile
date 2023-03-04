# build context at repo root: docker build -f Dockerfile ../..
FROM golang:1.16 AS builder

WORKDIR /app

COPY go.mod go.mod
COPY go.sum go.sum
COPY go go
COPY example/go example/go

WORKDIR /app/example/go

RUN CGO_ENABLED=0 go build -v -o .build/server .

# stage 2: production image
FROM alpine

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/example/go/.build/server /server

# Run the web service on container startup.
CMD ["/server"]
