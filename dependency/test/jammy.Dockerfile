FROM ubuntu:jammy

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]

WORKDIR /test
