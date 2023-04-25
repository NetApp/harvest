FROM golang:latest as builder

SHELL ["/bin/bash", "-c"]

ARG INSTALL_DIR=/opt/harvest
ARG BUILD_DIR=/opt/home
ARG VERSION=2.0
ARG RELEASE=nightly
ARG GIT_TOKEN
ARG ASUP_MAKE_TARGET=build

# Set the Current Working Directory inside the container
WORKDIR $BUILD_DIR

RUN mkdir -p $INSTALL_DIR $INSTALL_DIR/container/onePollerPerContainer $INSTALL_DIR/container/prometheus

COPY . .

RUN if [[ -n "$ASUP_MAKE_TARGET" && -n "$GIT_TOKEN" ]]; then \
make build asup VERSION=$VERSION RELEASE=$RELEASE ASUP_MAKE_TARGET=$ASUP_MAKE_TARGET GIT_TOKEN=$GIT_TOKEN ; \
else \
make build VERSION=$VERSION RELEASE=$RELEASE BIN_PLATFORM=linux ;\
fi

RUN cp -a $BUILD_DIR/harvest.yml $INSTALL_DIR/harvest.yml.example

RUN cp -aR bin $BUILD_DIR/conf $BUILD_DIR/grafana $BUILD_DIR/autosupport $BUILD_DIR/prom-stack.tmpl $INSTALL_DIR

RUN cp -a $BUILD_DIR/container/onePollerPerContainer/docker-compose.tmpl $INSTALL_DIR/container/onePollerPerContainer

RUN cp -aR $BUILD_DIR/container/prometheus $INSTALL_DIR/container/

FROM gcr.io/distroless/static-debian11:debug

ARG INSTALL_DIR=/opt/harvest
COPY --from=builder $INSTALL_DIR $INSTALL_DIR
WORKDIR $INSTALL_DIR

ENTRYPOINT ["bin/poller"]
