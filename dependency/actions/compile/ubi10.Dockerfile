FROM registry.access.redhat.com/ubi10/ubi

RUN dnf install -y \
      gcc \
      gcc-c++ \
      make \
      tar \
      gzip \
      curl-minimal

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
