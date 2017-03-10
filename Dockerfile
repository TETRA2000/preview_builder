FROM golang:1.6.4

ENV DOCKER_VERSION 1.12.4

RUN cd /tmp/ \
    && curl -sSL -O https://get.docker.com/builds/Linux/x86_64/docker-${DOCKER_VERSION}.tgz \
    && tar zxf docker-${DOCKER_VERSION}.tgz \
    && mkdir -p /usr/local/bin/ \
    && mv $(find -name 'docker' -type f) /usr/local/bin/ \
    && chmod +x /usr/local/bin/docker \
    && rm -rf /tmp/*

ADD ./src /opt/app/src
WORKDIR /opt/app
RUN GOPATH=$PWD go build src/main.go

#CMD ./main
CMD docker run hello-world