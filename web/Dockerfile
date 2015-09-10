FROM centos:7
RUN yum install -y git
COPY index.html /
COPY decap-linux-amd64 /
ENTRYPOINT ["/decap-linux-amd64"]
