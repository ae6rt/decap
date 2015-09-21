FROM centos:7

MAINTAINER Mark Petrovic <mspetrovic@gmail.com>

RUN yum install -y git bind-utils
RUN mkdir /static

COPY decap-linux-amd64 /
COPY static/index.html /static/

ENTRYPOINT ["/decap-linux-amd64"]
