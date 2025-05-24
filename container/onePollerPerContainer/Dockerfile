# GO_VERSION should be overridden by the build script via --build-arg GO_VERSION=$value
ARG GO_VERSION
FROM golang:${GO_VERSION} AS builder

SHELL ["/bin/bash", "-c"]

ARG INSTALL_DIR=/opt/harvest
ARG BUILD_DIR=/opt/home
ARG VERSION=2.0
ARG RELEASE=prod

# Set the Current Working Directory inside the container
WORKDIR $BUILD_DIR

RUN mkdir -p $INSTALL_DIR $INSTALL_DIR/container/onePollerPerContainer $INSTALL_DIR/cert

COPY . .

RUN make build VERSION=$VERSION RELEASE=$RELEASE BIN_PLATFORM=linux

RUN cp -a $BUILD_DIR/harvest.yml $INSTALL_DIR/harvest.yml.example \
    && cp -aR bin $BUILD_DIR/conf $INSTALL_DIR

FROM gcr.io/distroless/static-debian12:debug

ARG INSTALL_DIR=/opt/harvest
ENV HARVEST_DOCKER=yes

COPY --from=builder $INSTALL_DIR $INSTALL_DIR

WORKDIR $INSTALL_DIR

# Set GODEBUG environment variable based on build argument
ARG GODEBUG
ENV GODEBUG=${GODEBUG}

ENTRYPOINT ["bin/poller"]