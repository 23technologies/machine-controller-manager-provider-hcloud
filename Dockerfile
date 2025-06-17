#############      builder                                  #############
FROM golang:1.24.4 AS builder

WORKDIR /go/src/github.com/23technologies/machine-controller-manager-provider-hcloud
COPY . .
RUN make install

#############      base                                     #############
FROM gcr.io/distroless/static-debian12:nonroot AS base

WORKDIR /

#############      machine-controller               #############
FROM base AS machine-controller
WORKDIR /

COPY --from=builder /go/bin/machine-controller /machine-controller
ENTRYPOINT ["/machine-controller"]
