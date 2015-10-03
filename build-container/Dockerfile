FROM centos:7.1.1503

MAINTAINER Mark Petrovic <mspetrovic@gmail.com>

RUN yum update -y && yum install -y git tar
RUN mkdir -p /home/decap/workspace /build-artifacts /home/decap/buildscripts

COPY entrypoint.sh /usr/local/bin/
COPY bctool-linux-amd64 /usr/local/bin/bctool

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Bake in Dockerfile for reference purposes
COPY Dockerfile /opt/decap/Dockerfile
