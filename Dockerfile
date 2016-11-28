FROM golang:1.7
MAINTAINER Gennady Karev <pendolf666@gmail.com>

ADD . /go/src/github.com/maddevsio/http-agent

WORKDIR /go/src/github.com/maddevsio/http-agent

RUN go get -v && go build -v

CMD ["./http-agent"]
