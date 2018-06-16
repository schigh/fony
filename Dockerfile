FROM golang:1.10-alpine
MAINTAINER Steve High <steve.high@gmail.com>
COPY ./ /go/src/github.com/schigh/fony
WORKDIR /go/src/github.com/schigh/fony

RUN apk add --no-cache git bash
RUN go get -u github.com/kardianos/govendor
RUN govendor init
RUN govendor fetch github.com/labstack/echo@v1.4.4
RUN govendor fetch go.uber.org/zap@v1.8.0
RUN go build -o fony *.go
RUN mv fony /go/bin/fony

EXPOSE 80

CMD ["fony", "-f", "./fony.json"]

