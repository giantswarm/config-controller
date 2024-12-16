FROM gsoci.azurecr.io/giantswarm/alpine:3.21.0

RUN apk add --no-cache ca-certificates

ADD ./config-controller /config-controller

ENTRYPOINT ["/config-controller"]
