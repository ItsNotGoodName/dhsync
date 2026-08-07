FROM alpine

RUN apk add --no-cache tzdata

WORKDIR /config

ENTRYPOINT ["/usr/bin/dhsync", "daemon"]

ARG TARGETPLATFORM

COPY ./dist/${TARGETPLATFORM}/dhsync /usr/bin/dhsync
