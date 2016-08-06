# COMMAND FOR STATIC COMPILATION: CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' kvb.go
FROM alpine:3.4

ENV GOPATH /go

RUN apk update && apk add go git \
	&& mkdir -p /go/bin \
	&& mkdir -p /go/pkg \
	&& mkdir -p /go/src/github.com/pgray/kvb/


WORKDIR /go/src/github.com/pgray/kvb/
COPY . ./

RUN go get && CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' kvb.go

EXPOSE 8080
ENTRYPOINT ["./kvb"]
