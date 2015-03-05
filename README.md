UQAM Resultats Crawler
==============

[![Build Status](https://travis-ci.org/janicduplessis/resultscrawler.svg)](https://travis-ci.org/janicduplessis/resultscrawler)
[![GoDoc](https://godoc.org/github.com/janicduplessis/resultscrawler?status.svg)](https://godoc.org/github.com/janicduplessis/resultscrawler)
[![Coverage Status](https://coveralls.io/repos/janicduplessis/resultscrawler/badge.svg)](https://coveralls.io/r/janicduplessis/resultscrawler)

This application contains two executables, a crawler to fetch data from
the UQAM website and a webserver to access the data at any time.

The is a web client, built using Angularjs. An iOS and Androit app are also in developpement.

The project is currently running on https://results.jdupserver.com.

Prerequisites
---------------
Mandatory:

- [Go](http://golang.org/)
- [Node and npm](http://nodejs.org/)
- [mongodb](http://www.mongodb.org/)

Optional:

- An email smtp server

Installation
---------------
The recommended way to get the code is through the go get command.

        go get github.com/janicduplessis/resultscrawler

1.  Navigate to the root folder and install go dependencies.
You may need to install mercurial and bazaar to be able to download the dependencies.

        cd $GOPATH/src/github.com/janicduplessis/resultscrawler
        go install ./...

2.  If you don't already have bower installed globally, install it.

        npm install bower -g

3.  Install the webserver libraries using bower.

        cd webserver
        bower install

4.  Create config files.

    4.1  Crawler

    From the project root:

              cd crawler
              cp template.config.json crawler.config.json

    Edit the crawler.config.json file to reflect your server configuration.

    4.2  Webserver

    From the project root:

              cd webserver
              cp template.config.json webserver.config.json

    Edit the webserver.config.json file to reflect your server configuration.

Run the code
--------------
To run the crawler, from the project root:

        cd crawler
        go run crawler.go

To run the webserver, from the project root:

        cd webserver
        go run webserver.go


Hopefully everything worked!
