FROM centos:7.3.1611

MAINTAINER Mark Petrovic <mspetrovic@gmail.com>

RUN yum install -y git tar && yum clean all
RUN mkdir -p /home/decap/workspace /build-artifacts /home/decap/buildscripts

COPY entrypoint.sh /usr/local/bin/
COPY bctool-linux-amd64 /usr/local/bin/bctool

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Bake in Dockerfile for reference purposes
COPY Dockerfile /opt/decap/Dockerfile
