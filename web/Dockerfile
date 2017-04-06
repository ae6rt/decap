FROM centos:7.3.1611

MAINTAINER Mark Petrovic <mspetrovic@gmail.com>

RUN yum install -y git bind-utils

COPY static/ /static/
COPY build/dist/linux/decap /

ENTRYPOINT ["/decap"]

COPY Dockerfile /opt/decap/
