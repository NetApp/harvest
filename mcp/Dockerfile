ARG GO_VERSION
FROM golang:${GO_VERSION} AS builder

SHELL ["/bin/bash", "-c"]

ARG INSTALL_DIR=/opt/harvest-mcp
ARG BUILD_DIR=/opt/home
ARG VERSION=1.0.0

WORKDIR $BUILD_DIR

RUN mkdir -p $INSTALL_DIR

COPY mcp/ ./mcp/
COPY docs/ ./docs/

WORKDIR $BUILD_DIR/mcp

RUN make build VERSION=$VERSION

RUN mkdir -p $INSTALL_DIR/bin
RUN cp -a $BUILD_DIR/mcp/bin/harvest-mcp $INSTALL_DIR/bin/

RUN mkdir -p $INSTALL_DIR/docs
RUN cp -a $BUILD_DIR/docs/ontap-metrics.md $INSTALL_DIR/docs/ 2>/dev/null || true
RUN cp -a $BUILD_DIR/docs/cisco-switch-metrics.md $INSTALL_DIR/docs/ 2>/dev/null || true
RUN cp -a $BUILD_DIR/docs/storagegrid-metrics.md $INSTALL_DIR/docs/ 2>/dev/null || true

RUN mkdir -p $INSTALL_DIR/metadata
RUN cp -a $BUILD_DIR/mcp/metadata/*.json $INSTALL_DIR/metadata/ 2>/dev/null || true

RUN mkdir -p $INSTALL_DIR/prompts
RUN cp -a $BUILD_DIR/mcp/prompts/* $INSTALL_DIR/prompts/ 2>/dev/null || true

FROM gcr.io/distroless/static-debian12:debug

ARG INSTALL_DIR=/opt/harvest-mcp

COPY --from=builder $INSTALL_DIR $INSTALL_DIR

WORKDIR $INSTALL_DIR

ENTRYPOINT ["./bin/harvest-mcp", "start"]
