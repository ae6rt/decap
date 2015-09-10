FROM centos:7
RUN yum install -y git bind-utils
COPY index.html /
COPY decap-linux-amd64 /
ENTRYPOINT ["/decap-linux-amd64"]
