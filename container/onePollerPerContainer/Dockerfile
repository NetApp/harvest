# GO_VERSION should be overridden by the build script via --build-arg GO_VERSION=$value
ARG GO_VERSION
FROM mcr.microsoft.com/oss/go/microsoft/golang:${GO_VERSION}-fips-cbl-mariner2.0 AS builder

SHELL ["/bin/bash", "-c"]

ARG INSTALL_DIR=/opt/harvest
ARG BUILD_DIR=/opt/home
ARG VERSION=2.0
ARG RELEASE=nightly
ARG GIT_TOKEN
ARG ASUP_MAKE_TARGET=build
ARG FIPS=0

# Set the Current Working Directory inside the container
WORKDIR $BUILD_DIR

# Install necessary utilities
RUN tdnf install -y which && \
    tdnf clean all

RUN mkdir -p $INSTALL_DIR $INSTALL_DIR/container/onePollerPerContainer $INSTALL_DIR/container/prometheus $INSTALL_DIR/cert

COPY . .

RUN if [[ -n "$ASUP_MAKE_TARGET" && -n "$GIT_TOKEN" ]]; then \
make build asup VERSION=$VERSION RELEASE=$RELEASE ASUP_MAKE_TARGET=$ASUP_MAKE_TARGET GIT_TOKEN=$GIT_TOKEN FIPS=$FIPS; \
else \
make build VERSION=$VERSION RELEASE=$RELEASE BIN_PLATFORM=linux FIPS=$FIPS; \
fi

RUN cp -a $BUILD_DIR/harvest.yml $INSTALL_DIR/harvest.yml.example

RUN cp -aR bin $BUILD_DIR/conf $BUILD_DIR/grafana $BUILD_DIR/autosupport $BUILD_DIR/prom-stack.tmpl $INSTALL_DIR

RUN cp -a $BUILD_DIR/container/onePollerPerContainer/docker-compose.tmpl $INSTALL_DIR/container/onePollerPerContainer

RUN cp -aR $BUILD_DIR/container/prometheus $INSTALL_DIR/container/

FROM mcr.microsoft.com/cbl-mariner/base/core:2.0

ARG INSTALL_DIR=/opt/harvest
ENV HARVEST_DOCKER=yes
COPY --from=builder $INSTALL_DIR $INSTALL_DIR

RUN tdnf install -y ca-certificates && \
    tdnf clean all && \
    rm -rf /var/cache/tdnf

# A program built with systemcrypto always uses the OpenSSL library present on the system for crypto APIs
ENV OPENSSL_FIPS=1

WORKDIR $INSTALL_DIR

ENTRYPOINT ["bin/poller"]