FROM centos:7.1.1503

MAINTAINER Mark Petrovic <mspetrovic@gmail.com>

RUN yum update -y && yum install -y git
RUN curl -o /tmp/get-pip.py https://bootstrap.pypa.io/get-pip.py && python /tmp/get-pip.py && pip install awscli && rm /tmp/get-pip.py 
RUN mkdir -p /home/decap/workspace /build-artifacts /home/decap/buildscripts /opt/decap

COPY entrypoint.sh /opt/decap/bin/

ENTRYPOINT ["/opt/decap/bin/entrypoint.sh"]

# Bake in Dockerfile for reference purposes
COPY Dockerfile /opt/decap/
