#############      builder                                  #############
FROM eu.gcr.io/gardener-project/3rd/golang:1.18.4 AS builder

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
