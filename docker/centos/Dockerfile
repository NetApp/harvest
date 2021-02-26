FROM centos:centos9
MAINTAINER vachagan@netapp.com

RUN \
    yum update -y && \
    yum install -y && \
        tar \
        gcc \
        libc6-dev && \
    yum install -y /harvest/harvest*.rpm && \
    harvest status > /harvest/conf/pollers.txt && \
    mv /etc/harvest/* /harvest/conf/ &&
    export HARVEST_CONF=/harvest/conf &&
    export HARVEST_LOGS=/harvest/log &&
    export HARVEST_PIDS=/harvest/pid &&
    echo "set up docker harvest complete"

CMD ["harvest", "start" "$HARVEST_POLLER"]

