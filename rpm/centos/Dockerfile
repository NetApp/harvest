FROM centos:centos7
MAINTAINER vachagan@netapp.com

ENV GOPATH /go
ENV GOVERSION 1.19
ENV GOURL https://go.dev/dl/go$GOVERSION.linux-amd64.tar.gz

RUN \
#  yum update -y && \
#  yum install -y \
#    epel-release && \
  yum update -y && \
  yum install -y \
    rpm-build \
    tar \
    wget \
    make \
    which \
    sudo \
    gcc \
    libc6-dev && \
  yum clean all && \
  curl -fsSL "$GOURL" -o golang.tar.gz && \
  tar -C /usr/local -xzf golang.tar.gz && \
  rm golang.tar.gz && \
  mkdir -p "$GOPATH/src" "$GOPATH/bin" && \
  chmod -R 777 "$GOPATH" && \
  ln -s "/usr/local/go/bin/go" "/usr/local/bin/go" && \
  echo "installed go $GOVERSION"
CMD ["bash", "/tmp/src/rpm/build-rpm.sh"]

