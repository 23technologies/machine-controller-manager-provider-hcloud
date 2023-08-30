#############      builder                                  #############
FROM golang:1.20.2 AS builder

ENV BINARY_PATH=/go/bin
WORKDIR /go/src/github.com/23technologies/machine-controller-manager-provider-hcloud

COPY . .
RUN hack/build.sh

#############      base                                     #############
FROM gcr.io/distroless/static-debian11:nonroot as base

WORKDIR /

#############      machine-controller               #############
FROM base AS machine-controller
LABEL org.opencontainers.image.source="https://github.com/23technologies/machine-controller-manager-provider-hcloud"

COPY --from=builder /go/bin/machine-controller /machine-controller
ENTRYPOINT ["/machine-controller"]
