FROM gcr.io/distroless/static-debian12:debug

ARG TARGETARCH
ARG INSTALL_DIR=/opt/harvest
ENV HARVEST_DOCKER=yes

COPY docker-build/bin.${TARGETARCH}/ $INSTALL_DIR/bin/
COPY docker-build/autosupport.${TARGETARCH}/ $INSTALL_DIR/autosupport/

COPY harvest.yml $INSTALL_DIR/harvest.yml.example
COPY conf/ $INSTALL_DIR/conf/
COPY grafana/ $INSTALL_DIR/grafana/
COPY prom-stack.tmpl $INSTALL_DIR/prom-stack.tmpl
COPY container/onePollerPerContainer/docker-compose.tmpl $INSTALL_DIR/container/onePollerPerContainer/docker-compose.tmpl
COPY container/prometheus/ $INSTALL_DIR/container/prometheus/

WORKDIR $INSTALL_DIR

ARG GODEBUG
ENV GODEBUG=${GODEBUG}

ENTRYPOINT ["bin/poller"]