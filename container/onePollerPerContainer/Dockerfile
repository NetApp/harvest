# GO_VERSION should be overridden by the build script via --build-arg GO_VERSION=$value
ARG GO_VERSION
FROM golang:${GO_VERSION} AS builder

SHELL ["/bin/bash", "-c"]

ARG ZIG_VERSION=0.13.0
ARG INSTALL_DIR=/opt/harvest
ARG BUILD_DIR=/opt/home
ARG VERSION=2.0
ARG RELEASE=nightly
ARG ASUP_MAKE_TARGET=build
ARG FIPS=0

# Set the Current Working Directory inside the container
WORKDIR $BUILD_DIR

RUN apt-get update && apt-get install -y xz-utils

# Install Zig if FIPS is enabled
RUN if [ "$FIPS" -eq "1" ]; then \
    wget https://ziglang.org/download/${ZIG_VERSION}/zig-linux-x86_64-${ZIG_VERSION}.tar.xz \
    && tar -xf zig-linux-x86_64-${ZIG_VERSION}.tar.xz \
    && mv zig-linux-x86_64-${ZIG_VERSION} /usr/local/zig; \
    fi

RUN mkdir -p $INSTALL_DIR $INSTALL_DIR/container/onePollerPerContainer $INSTALL_DIR/container/prometheus $INSTALL_DIR/cert

COPY . .

RUN --mount=type=secret,id=git_token \
    if [ "$FIPS" -eq "1" ]; then \
    export PATH="/usr/local/zig:${PATH}"; \
    fi && \
    if [[ -n "$ASUP_MAKE_TARGET" && -f "/run/secrets/git_token" ]]; then \
    make build asup VERSION=$VERSION RELEASE=$RELEASE ASUP_MAKE_TARGET=$ASUP_MAKE_TARGET GIT_TOKEN=$(cat /run/secrets/git_token) FIPS=$FIPS ; \
    else \
    make build VERSION=$VERSION RELEASE=$RELEASE BIN_PLATFORM=linux FIPS=$FIPS ;\
    fi

RUN cp -a $BUILD_DIR/harvest.yml $INSTALL_DIR/harvest.yml.example \
    && cp -aR bin $BUILD_DIR/conf $BUILD_DIR/grafana $BUILD_DIR/autosupport $BUILD_DIR/prom-stack.tmpl $INSTALL_DIR \
    && cp -a $BUILD_DIR/container/onePollerPerContainer/docker-compose.tmpl $INSTALL_DIR/container/onePollerPerContainer \
    && cp -aR $BUILD_DIR/container/prometheus $INSTALL_DIR/container/

# Create directories and conditionally copy .so files if FIPS is enabled
RUN mkdir -p /temp_libs/x86_64-linux-gnu /temp_libs/lib64 && \
    if [ "$FIPS" -eq "1" ]; then \
    cp /lib/x86_64-linux-gnu/libresolv.so.2 /temp_libs/x86_64-linux-gnu/ && \
    cp /lib/x86_64-linux-gnu/libpthread.so.0 /temp_libs/x86_64-linux-gnu/ && \
    cp /lib/x86_64-linux-gnu/libc.so.6 /temp_libs/x86_64-linux-gnu/ && \
    cp /lib64/ld-linux-x86-64.so.2 /temp_libs/lib64/; \
    fi

FROM gcr.io/distroless/static-debian12:debug

ARG INSTALL_DIR=/opt/harvest
ENV HARVEST_DOCKER=yes
COPY --from=builder $INSTALL_DIR $INSTALL_DIR

# Conditionally copy .so files from builder stage if they exist
COPY --from=builder /temp_libs/x86_64-linux-gnu/ /lib/x86_64-linux-gnu/
COPY --from=builder /temp_libs/lib64/ /lib64/

WORKDIR $INSTALL_DIR

ENTRYPOINT ["bin/poller"]