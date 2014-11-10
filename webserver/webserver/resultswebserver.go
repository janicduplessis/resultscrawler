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
	Webserver *Webserver
	UserStore lib.UserStore
	Crypto    lib.Crypto
}

func NewResultsWebserver(ws *Webserver, userStore lib.UserStore, crypto lib.Crypto) *ResultsWebserver {
	wsHandler := &ResultsWebserver{
		Webserver: ws,
		UserStore: userStore,
		Crypto:    crypto,
	}

	ws.Router.HandleFunc("/", ws.MakeHandler(false, wsHandler.HomeHandler)).Methods("GET")
	ws.Router.HandleFunc("/about", ws.MakeHandler(false, wsHandler.AboutHandler)).Methods("GET")

	return wsHandler
}

func (handler *ResultsWebserver) HomeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	model := homeModel{
		Page{
			PageID: "home",
			Title:  "Home",
		},
	}

	err := handler.Webserver.Templates.ExecuteTemplate(w, "index", &model)
	if err != nil {
		handler.Webserver.Error(w, err)
		return
	}
}

func (handler *ResultsWebserver) AboutHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	model := aboutModel{
		Page{
			PageID: "home",
			Title:  "Home",
		},
	}

	err := handler.Webserver.Templates.ExecuteTemplate(w, "about", &model)
	if err != nil {
		handler.Webserver.Error(w, err)
		return
	}
}
