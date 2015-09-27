FROM centos:7.1.1503

MAINTAINER Mark Petrovic <mspetrovic@gmail.com>

RUN yum update -y && yum install -y git bind-utils

COPY static/index.html /static/
COPY decap-linux-amd64 /

ENTRYPOINT ["/decap-linux-amd64"]

COPY Dockerfile /opt/decap/
