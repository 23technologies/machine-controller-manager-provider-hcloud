#############      builder                                  #############
FROM eu.gcr.io/gardener-project/3rd/golang:1.16.2 AS builder

WORKDIR /go/src/github.com/23technologies/machine-controller-manager-provider-hcloud
COPY . .

RUN ./hack/build.sh

#############      base                                     #############
FROM eu.gcr.io/gardener-project/3rd/alpine:3.13.2 as base

RUN apk add --update bash curl tzdata
WORKDIR /

#############      machine-controller               #############
FROM base AS machine-controller

COPY --from=builder /go/src/github.com/23technologies/machine-controller-manager-provider-hcloud/bin/rel/machine-controller /machine-controller
ENTRYPOINT ["/machine-controller"]
