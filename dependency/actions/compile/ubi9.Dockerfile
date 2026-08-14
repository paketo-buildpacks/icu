FROM registry.access.redhat.com/ubi9/ubi

RUN dnf install -y \
      gcc \
      gcc-c++ \
      make \
      tar \
      gzip \
      curl-minimal

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
