FROM golang:1.23.5

WORKDIR ${GOPATH}/avito_pvz/
COPY . ${GOPATH}/avito_pvz/

RUN go build -o /build ./cmd \
    && go clean -cache -modcache

EXPOSE 8080

CMD ["/build"]