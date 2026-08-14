FROM registry.access.redhat.com/ubi8/ubi

RUN dnf install -y \
      gcc \
      gcc-c++ \
      make \
      tar \
      gzip \
      curl

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
