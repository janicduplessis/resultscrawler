package webserver

import (
	"net/http"

	"code.google.com/p/go.net/context"

	"github.com/janicduplessis/resultscrawler/lib"
)

type Webserver interface {
	Redirect(w http.ResponseWriter, r *http.Request, url string)
	Log(msg string)
	Error(w http.ResponseWriter, err error)
	StartSession(ctx context.Context, w http.ResponseWriter, r *http.Request, user *lib.User) error
	EndSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	AddHandler(url string, authenticated bool, fn func(context.Context, http.ResponseWriter, *http.Request))
}
