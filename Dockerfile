# COMMAND FOR STATIC COMPILATION: CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' kvb.go

FROM scratch
COPY kvb /kvb
COPY templates/* /templates/
EXPOSE 8080
ENTRYPOINT ["/kvb"]
