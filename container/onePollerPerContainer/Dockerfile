# GO_VERSION should be overridden by the build script via --build-arg GO_VERSION=$value
ARG GO_VERSION
FROM golang:${GO_VERSION} AS builder

SHELL ["/bin/bash", "-c"]

ARG ZIG_VERSION=0.13.0
ARG INSTALL_DIR=/opt/harvest
ARG BUILD_DIR=/opt/home
ARG VERSION=2.0
ARG RELEASE=nightly
ARG GIT_TOKEN
ARG ASUP_MAKE_TARGET=build
ARG FIPS=0

# Set the Current Working Directory inside the container
WORKDIR $BUILD_DIR

RUN apt-get update && apt-get install -y xz-utils

# Install Zig if FIPS is enabled and set PATH
RUN if [ "$FIPS" -eq "1" ]; then \
    wget https://ziglang.org/download/${ZIG_VERSION}/zig-linux-x86_64-${ZIG_VERSION}.tar.xz \
    && tar -xf zig-linux-x86_64-${ZIG_VERSION}.tar.xz \
    && mv zig-linux-x86_64-${ZIG_VERSION} /usr/local/zig \
    && ln -s /usr/local/zig/zig /usr/local/bin/zig \
    && export PATH="/usr/local/zig:${PATH}"; \
    fi

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

FROM gcr.io/distroless/base-debian12:debug

ARG INSTALL_DIR=/opt/harvest
ENV HARVEST_DOCKER=yes
COPY --from=builder $INSTALL_DIR $INSTALL_DIR
WORKDIR $INSTALL_DIR

ENTRYPOINT ["bin/poller"]