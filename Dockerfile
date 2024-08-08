FROM gsoci.azurecr.io/giantswarm/alpine:3.20.2

RUN apk add --no-cache ca-certificates

ADD ./config-controller /config-controller

ENTRYPOINT ["/config-controller"]
