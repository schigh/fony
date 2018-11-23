FROM golang:alpine AS builder
WORKDIR /fony/cmd
ADD ./ /fony
RUN apk add --no-cache git bash
RUN mkdir /build
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/fony .

FROM gcr.io/distroless/base
COPY --from=builder  /build/fony /
EXPOSE 80
CMD ["/fony"]
