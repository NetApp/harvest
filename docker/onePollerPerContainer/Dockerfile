FROM golang:1.16-alpine as golangAlpine

RUN apk add curl git && apk add --virtual build-dependencies build-base gcc

ARG INSTALL_DIR=/opt/harvest
ARG BUILD_DIR=/opt/home
ARG VERSION=2.0
ARG RELEASE=nightly
ARG ENABLE_ASUP=N
ARG GIT_TOKEN=
ARG ASUP_MAKE_TARGET=build

RUN echo $VERSION
RUN echo $RELEASE

RUN if [[ "$ENABLE_ASUP" == "Y" || "$ENABLE_ASUP" == "y" ]] && [ -z "$GIT_TOKEN" ];then \
  >&2 echo "GIT_TOKEN is required when ENABLE_ASUP is Y!" ;\
  exit 1; \
  fi

# Set the Current Working Directory inside the container
WORKDIR $BUILD_DIR

RUN mkdir -p $INSTALL_DIR

COPY . .

RUN if [[ "$ENABLE_ASUP" != "Y" && "$ENABLE_ASUP" != "y" ]]; then \
make build VERSION=$VERSION RELEASE=$RELEASE ; \
else \
make asup build VERSION=$VERSION RELEASE=$RELEASE GIT_TOKEN=$GIT_TOKEN ASUP_MAKE_TARGET=$ASUP_MAKE_TARGET ;\
fi

RUN cp -a $BUILD_DIR/harvest.yml $INSTALL_DIR/harvest.yml

RUN cp -aR bin $BUILD_DIR/conf $BUILD_DIR/docs $BUILD_DIR/grafana $INSTALL_DIR

FROM alpine:latest

RUN apk add bash

ARG INSTALL_DIR=/opt/harvest
COPY --from=golangAlpine $INSTALL_DIR $INSTALL_DIR
WORKDIR $INSTALL_DIR

ENTRYPOINT ["bin/poller"]
#CMD [ "--config", "harvest.yml", "--loglevel", "2"]
