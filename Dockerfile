FROM golang as build
WORKDIR /go/src/consul-demo-service
COPY go.mod .
COPY go.sum .
COPY consul-demo-service.go .
ENV CGO_ENABLED=0
RUN go build -a -ldflags \
  '-extldflags "-static"'

FROM busybox
COPY --from=build /go/src/consul-demo-service/consul-demo-service /
CMD ["/consul-demo-service"]
EXPOSE 80
