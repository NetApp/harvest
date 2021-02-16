FROM centos:centos7
MAINTAINER vachagan@netapp.com

ENV GOPATH /go
ENV GOVERSION 1.15.8
ENV GOURL https://golang.org/dl/go$GOVERSION.linux-amd64.tar.gz

RUN \
#  yum update -y && \
#  yum install -y \
#    epel-release && \
  yum update -y && \
  yum install -y \
    rpm-build \
    tar \
    wget \
    gcc \
    libc6-dev && \
  yum clean all && \
  curl -fsSL "$GOURL" -o golang.tar.gz && \
  #echo "$GOSHA golang.tar.gz" | sha256sum -c - && \
  tar -C /usr/local -xzf golang.tar.gz && \
  rm golang.tar.gz && \
  mkdir -p "$GOPATH/src" "$GOPATH/bin" && \
  chmod -R 777 "$GOPATH" && \
  ln -s "/usr/local/go/bin/go" "/usr/local/bin/go" && \
  echo "installed go $GOVERSION"
CMD ["sh", "/tmp/xapp/cmd/docker-build/build-rpm.sh"]

