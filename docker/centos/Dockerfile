FROM centos:centos7
ENV HARVEST_PACK harvest-2.0.1-*.x86_64.rpm
COPY $HARVEST_PACK /

RUN \
    #yum update -y && \
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
    #harvest config welcome 
    echo "export HARVEST_CONF=/harvest/" > /opt/harvest/sources.sh && \
    echo "export HARVEST_LOGS=/harvest/log" >> /opt/harvest/sources.sh && \
    echo "export HARVEST_PIDS=/harvest/pid" >> /opt/harvest/sources.sh && \
    echo "set custom path variables:" && \
    cat /opt/harvest/sources.sh && \
    ls /opt/harvest/ -a && \
    echo "docker harvest complete"

#CMD ["harvest", "start", "jamaica", "--loglevel", "1"]
#CMD ["harvest", "$HARVEST_CMD", "$HARVEST_ARG", "--config", "/harvest.yml"]
CMD harvest "$HARVEST_CMD" "$HARVEST_ARG"
#CMD ["harvest", "config", "welcome"]
#CMD ["whiptail", "--yesno", "continue?", "0", "0"]
#CMD which whiptail
