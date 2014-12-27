FROM golang:1.4-wheezy

RUN go get github.com/janicduplessis/resultscrawler/crawler
WORKDIR $GOPATH/src/github.com/janicduplessis/resultscrawler/crawler

CMD []
ENTRYPOINT ["$GOPATH/bin/crawler"]
