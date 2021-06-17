FROM golang:1.16-alpine as golangAlpine

RUN apk add curl && apk add --virtual build-dependencies build-base gcc

ARG INSTALL_DIR=/opt/harvest
ARG BUILD_DIR=/opt/home
ARG VERSION=2.0
ARG RELEASE=nightly

RUN echo $VERSION
RUN echo $RELEASE

# Set the Current Working Directory inside the container
WORKDIR $BUILD_DIR

RUN mkdir -p $INSTALL_DIR

COPY . .

RUN make build VERSION=$VERSION RELEASE=$RELEASE

RUN cp -a $BUILD_DIR/harvest.yml $INSTALL_DIR/harvest.yml

RUN cp -aR bin $BUILD_DIR/conf $BUILD_DIR/docs $BUILD_DIR/grafana $BUILD_DIR/docker/onePollerPerContainer/docker-entrypoint.sh $INSTALL_DIR && \
   chmod +x $INSTALL_DIR/docker-entrypoint.sh

FROM alpine:latest

# Work around https://github.com/mirror/busybox/issues/21
RUN apk add wget

RUN wget -O /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_x86_64 && \
    chmod +x /usr/local/bin/dumb-init

RUN apk add bash

ARG INSTALL_DIR=/opt/harvest
COPY --from=golangAlpine $INSTALL_DIR $INSTALL_DIR
WORKDIR $INSTALL_DIR

VOLUME     [ "/var/log/harvest" ]

ENTRYPOINT ["/usr/local/bin/dumb-init", "--", "./docker-entrypoint.sh"]
#CMD [ "--config", "harvest.yml", "--loglevel", "2"]
