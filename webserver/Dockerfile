FROM google/golang

RUN go get github.com/janicduplessis/resultscrawler/webserver
WORKDIR /gopath/src/github.com/janicduplessis/resultscrawler/webserver

CMD []
ENTRYPOINT ["/gopath/bin/webserver"]

EXPOSE 8080