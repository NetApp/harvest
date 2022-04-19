FROM golang:alpine as golangAlpine

RUN apk add curl && apk add --virtual build-dependencies build-base gcc git bash

ARG INSTALL_DIR=/opt/harvest
ARG BUILD_DIR=/opt/home
ARG VERSION=2.0
ARG RELEASE=nightly
ARG GIT_TOKEN
ARG ASUP_MAKE_TARGET=build

# Set the Current Working Directory inside the container
WORKDIR $BUILD_DIR

RUN mkdir -p $INSTALL_DIR

COPY . .

RUN if [[ -n "$ASUP_MAKE_TARGET" && -n "$GIT_TOKEN" ]]; then \
make build asup VERSION=$VERSION RELEASE=$RELEASE ASUP_MAKE_TARGET=$ASUP_MAKE_TARGET GIT_TOKEN=$GIT_TOKEN ; \
else \
make build VERSION=$VERSION RELEASE=$RELEASE BIN_PLATFORM=alpine ;\
fi

RUN cp -a $BUILD_DIR/harvest.yml $INSTALL_DIR/harvest.yml

RUN cp -aR bin $BUILD_DIR/conf $BUILD_DIR/docs $BUILD_DIR/grafana $BUILD_DIR/autosupport $INSTALL_DIR

FROM alpine@sha256:a777c9c66ba177ccfea23f2a216ff6721e78a662cd17019488c417135299cd89

RUN apk add bash

ARG INSTALL_DIR=/opt/harvest
COPY --from=golangAlpine $INSTALL_DIR $INSTALL_DIR
WORKDIR $INSTALL_DIR

ENTRYPOINT ["bin/poller"]
#CMD [ "--config", "harvest.yml", "--loglevel", "2"]
