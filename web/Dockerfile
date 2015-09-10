FROM centos:7
COPY decap-linux-amd64 /
COPY index.html /
RUN yum install -y git
ENTRYPOINT ["/decap-linux-amd64"]
