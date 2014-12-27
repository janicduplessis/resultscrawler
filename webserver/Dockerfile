FROM janicduplessis/golang-bower

RUN go get github.com/janicduplessis/resultscrawler/webserver
WORKDIR $GOPATH/src/github.com/janicduplessis/resultscrawler/webserver
RUN \
  cd $GOPATH/src/github.com/janicduplessis/resultscrawler/webserver && \
  bower --allow-root --config.interactive=false install

CMD []
ENTRYPOINT ["$GOPATH/bin/webserver"]

EXPOSE 8080
