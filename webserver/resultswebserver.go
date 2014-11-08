package webserver

import (
	"net/http"

	"code.google.com/p/go.net/context"

	"github.com/janicduplessis/resultscrawler/lib"
)

const (
	urlHome string = "/home"
)

type ResultsWebserver struct {
	Webserver Webserver
	UserStore lib.UserStore
	Crypto    lib.Crypto
}

func NewResultsWebserver(ws Webserver, userStore lib.UserStore, crypto lib.Crypto) *ResultsWebserver {
	wsHandler := &ResultsWebserver{
		Webserver: ws,
		UserStore: userStore,
		Crypto:    crypto,
	}

	ws.AddHandler(urlHome, false, wsHandler.HomeHandler)

	return wsHandler
}

func (handler *ResultsWebserver) HomeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}
