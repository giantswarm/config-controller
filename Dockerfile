FROM alpine:3.16.0

RUN apk add --no-cache ca-certificates

ADD ./config-controller /config-controller

ENTRYPOINT ["/config-controller"]
