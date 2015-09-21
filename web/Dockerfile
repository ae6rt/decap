FROM centos:7.1.1503

MAINTAINER Mark Petrovic <mspetrovic@gmail.com>

RUN yum update -y && yum install -y git bind-utils
RUN mkdir /static

COPY decap-linux-amd64 /
COPY static/index.html /static/

ENTRYPOINT ["/decap-linux-amd64"]
