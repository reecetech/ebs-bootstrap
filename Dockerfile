# syntax=docker/dockerfile:1
FROM golang:1.21

# Build a static binary
ENV CGO_ENABLED=0

# Set destination for COPY
WORKDIR /app

# Copy go source code
COPY ./ ./

# Install dependencies
RUN go mod download

# Build application
RUN go build cmd/ebs-bootstrap.go

# Test application
RUN go test ./...

# ebs-bootstrap cannot run in docker as it needs to interact
# with the raw devices of the host. Therefore docker must be
# exclusively used to build the binary in host architecture agnostic manner
CMD ["tail", "-f", "/dev/null"]
