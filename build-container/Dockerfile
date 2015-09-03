FROM centos:7

RUN curl -o /tmp/get-pip.py https://bootstrap.pypa.io/get-pip.py 
RUN python /tmp/get-pip.py 
RUN pip install awscli

RUN mkdir -p /root/.aws
RUN mkdir -p /home/decap/workspace
RUN mkdir -p /home/decap/buildscripts
COPY build.sh /opt/decap/bin/

ENTRYPOINT ["/opt/decap/bin/build.sh"]

# Bake in Dockerfile for reference purposes
RUN mkdir -p /opt/decap
COPY Dockerfile /opt/decap/
