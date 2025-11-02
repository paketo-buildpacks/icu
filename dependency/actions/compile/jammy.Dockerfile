FROM ubuntu:jammy

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get -y install curl build-essential

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
