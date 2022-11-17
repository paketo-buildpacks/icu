FROM ubuntu:bionic

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]

WORKDIR /test
