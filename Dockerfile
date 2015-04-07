FROM scratch
COPY kvb /kvb
COPY templates/* /templates/
EXPOSE 8080
ENTRYPOINT ["/kvb"]
