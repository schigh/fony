FROM golang:1.9-alpine
COPY ./ /go/src/github.com/schigh/fony
WORKDIR /go/src/github.com/schigh/fony

RUN apk add --no-cache git bash
RUN go get -u github.com/sirupsen/logrus
RUN go get -u goji.io
RUN GOBIN=/go/bin go install fony.go

