FROM golang:latest

RUN yes | apt-get update
RUN yes | apt-get upgrade
RUN yes | apt-get install rake

RUN mkdir -p /go/src/github.com/appcanary/agent

RUN mkdir -p /root/.ssh
RUN touch /root/.ssh/known_hosts
RUN ssh-keyscan github.com >> /root/.ssh/known_hosts

ADD . /go/src/github.com/appcanary/agent
WORKDIR /go/src/github.com/appcanary/agent

RUN go get -t -d -v ./...
