FROM alpine

WORKDIR /config

ENTRYPOINT ["/usr/bin/dhsync", "--daemon"]

ARG TARGETPLATFORM

COPY ./dist/${TARGETPLATFORM}/dhsync /usr/bin/dhsync
