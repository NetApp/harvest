# GO_VERSION should be overridden by the build script via --build-arg GO_VERSION=$value
ARG GO_VERSION
FROM golang:${GO_VERSION} AS builder

SHELL ["/bin/bash", "-c"]

ARG INSTALL_DIR=/opt/harvest
ARG BUILD_DIR=/opt/home
ARG VERSION=2.0
ARG RELEASE=nightly
ARG ASUP_MAKE_TARGET=build

# Set the Current Working Directory inside the container
WORKDIR $BUILD_DIR

RUN mkdir -p $INSTALL_DIR $INSTALL_DIR/container/onePollerPerContainer $INSTALL_DIR/container/prometheus $INSTALL_DIR/cert

COPY . .

RUN --mount=type=secret,id=git_token \
    if [[ -n "$ASUP_MAKE_TARGET" && -f "/run/secrets/git_token" ]]; then \
    make build asup VERSION=$VERSION RELEASE=$RELEASE ASUP_MAKE_TARGET=$ASUP_MAKE_TARGET GIT_TOKEN=$(cat /run/secrets/git_token); \
    else \
    make build VERSION=$VERSION RELEASE=$RELEASE BIN_PLATFORM=linux; \
    fi

RUN cp -a $BUILD_DIR/harvest.yml $INSTALL_DIR/harvest.yml.example \
    && cp -aR bin $BUILD_DIR/conf $BUILD_DIR/grafana $BUILD_DIR/autosupport $BUILD_DIR/prom-stack.tmpl $INSTALL_DIR \
    && cp -a $BUILD_DIR/container/onePollerPerContainer/docker-compose.tmpl $INSTALL_DIR/container/onePollerPerContainer \
    && cp -aR $BUILD_DIR/container/prometheus $INSTALL_DIR/container/

FROM gcr.io/distroless/static-debian12:debug

ARG INSTALL_DIR=/opt/harvest
ENV HARVEST_DOCKER=yes

COPY --from=builder $INSTALL_DIR $INSTALL_DIR

WORKDIR $INSTALL_DIR

# Set GODEBUG environment variable based on build argument
ARG GODEBUG
ENV GODEBUG=${GODEBUG}

ENTRYPOINT ["bin/poller"]