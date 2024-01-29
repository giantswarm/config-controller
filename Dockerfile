FROM alpine:3.19.1

RUN apk add --no-cache ca-certificates

ADD ./config-controller /config-controller

ENTRYPOINT ["/config-controller"]
