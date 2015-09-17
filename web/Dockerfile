FROM centos:7
RUN yum install -y git bind-utils
COPY decap-linux-amd64 /
RUN mkdir /static
COPY static/index.html /static/
ENTRYPOINT ["/decap-linux-amd64"]
