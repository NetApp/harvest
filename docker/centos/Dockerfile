FROM centos:centos7
ENV HARVEST_PACK harvest-2.0.1-*.x86_64.rpm
COPY $HARVEST_PACK /

RUN \
    yum update -y && \
    yum install -y \
        tar \
        gcc \
        newt \
        openssl \
        which \
        libc6-dev && \
    yum clean all && \
    rpm -ivh "$HARVEST_PACK" && \
    rm -f /"$HARVEST_PACK" && \
    echo "export HARVEST_CONF=/harvest/" > /opt/harvest/sources.sh && \
    echo "export HARVEST_LOGS=/harvest/log" >> /opt/harvest/sources.sh && \
    echo "export HARVEST_PIDS=/harvest/pid" >> /opt/harvest/sources.sh && \
    echo "docker harvest complete"

CMD harvest "$HARVEST_CMD" "$HARVEST_ARG"
