FROM golang:1.4-wheezy

# Install wget
RUN \
  apt-get update && \
  apt-get install -y wget && \
  rm -rf /var/lib/apt/lists/*

# Install Node.js and npm
RUN \
  cd /tmp && \
  wget http://nodejs.org/dist/node-latest.tar.gz && \
  tar xvzf node-latest.tar.gz && \
  rm -f node-latest.tar.gz && \
  cd node-v* && \
  ./configure && \
  CXX="g++ -Wno-unused-local-typedefs" make && \
  CXX="g++ -Wno-unused-local-typedefs" make install && \
  cd /tmp && \
  rm -rf /tmp/node-v* && \
  npm install -g npm && \
  echo -e '\n# Node.js\nexport PATH="node_modules/.bin:$PATH"' >> /root/.bashrc

RUN go get github.com/janicduplessis/resultscrawler/webserver
WORKDIR /usr/src/github.com/janicduplessis/resultscrawler/webserver
RUN \
  cd /usr/src/github.com/janicduplessis/resultscrawler/webserver && \
  npm install -g bower && \
  bower --allow-root --config.interactive=false install

CMD []
ENTRYPOINT ["/usr/bin/webserver"]

EXPOSE 8080
