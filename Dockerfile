# Dynamic Builds
ARG BUILDER_IMAGE=golang:1.19-bookworm
ARG FINAL_IMAGE=debian:bookworm-slim

# Build Stage
FROM --platform=${BUILDPLATFORM} ${BUILDER_IMAGE} AS builder

# Build Args
ARG GIT_REVISION=""

# Ensure ca-certificates are up to date
RUN update-ca-certificates

# Use modules for dependencies
WORKDIR $GOPATH/src/github.com/trisacrypto/courier

COPY go.mod .
COPY go.sum .

ENV CGO_ENABLED=0
ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

# Copy package
COPY . .

# Build the binary
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o /go/bin/courier -ldflags="-X 'github.com/trisacrypto/courier/pkg.GitVersion=${GIT_REVISION}'" ./cmd/courier

# Final Stage
FROM --platform=${BUILDPLATFORM} ${FINAL_IMAGE} AS final

LABEL maintainer="TRISA <info@trisa.io>"
LABEL description="Courier TSP Certificate Delivery Service"

# Ensure ca-certificates are up to date
RUN set -x && apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates sqlite3 && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /go/bin/courier /usr/local/bin/courier

# Create a user so that we don't run as root
RUN groupadd -r courier && useradd -m -r -g courier courier
USER courier

CMD [ "/usr/local/bin/courier", "serve" ]