FROM debian:debian9
MAINTAINER vachagan@netapp.com

RUN \
  apt-get update && \
  apt-get -y upgrade && \
  apt-get -y install build-essential gcc libc6-dev golang && \
  apt-get clean

CMD ["sh", "/tmp/src/cmd/deb/build-deb.sh"]

