FROM centos:centos7
COPY harvest-2.0.1-10.x86_64.rpm /

RUN \
    #yum update -y && \
    yum install -y \
        tar \
        gcc \
        dialog \
        newt \
        openssl \
        python3 \
        python3-pip \
        libc6-dev && \
    yum clean all && \
    pip3 install pyyaml && \
    rpm -ivh /harvest-2.0.1-10.x86_64.rpm && \
    rm -f /harvest-2.0.1-10.x86_64.rpm && \
    #harvest config welcome 
    export HARVEST_CONF="/harvest/conf" && \
    export HARVEST_LOGS="/harvest/log" && \
    export HARVEST_PIDS="/harvest/pid" && \
    echo "docker harvest complete"

CMD ["harvest", "config", "welcome"] # $HARVEST_CMD $HARVEST_ARG]
