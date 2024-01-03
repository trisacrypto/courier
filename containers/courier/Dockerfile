# Dynamic Builds
ARG BUILDER_IMAGE=golang:1.21-buster
ARG FINAL_IMAGE=debian:buster-slim

# Build Stage
FROM ${BUILDER_IMAGE} AS builder

ARG GIT_REVISION=""

# Ensure ca-certificates are up to date
RUN update-ca-certificates

# Use modeules for dependencies
WORKDIR $GOPATH/src/github.com/trisacrypto/courier
COPY go.mod .
COPY go.sum .

ENV CGO_ENABLED=0
ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

# Copy only what is needed for the build
COPY cmd ./cmd
COPY pkg ./pkg

# Build the binary
RUN go build -v -o /go/bin/courier -ldflags="-X 'github.com/trisacrypto/courier/pkg.GitVersion=$GIT_REVISION'" ./cmd/courier

# Final Stage
FROM ${FINAL_IMAGE} AS final

LABEL maintainer="TRISA <info@trisa.io>"
LABEL description="Certificate Delivery Service"

# Ensure ca-certificates are up to date
RUN set -x && apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /go/bin/courier /usr/local/bin/courier

CMD [ "/usr/local/bin/courier", "serve" ]