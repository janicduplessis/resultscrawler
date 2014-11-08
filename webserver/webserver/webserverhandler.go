package webserver

import (
	"encoding/gob"
	"net/http"

	"code.google.com/p/go.net/context"
	"github.com/gorilla/sessions"

	"github.com/janicduplessis/resultscrawler/lib"
)

const (
	sessionKey  = "AWDoijwapajwdvz23423oaawc"
	sessionName = "ct-session"
	ctxUser     = "User"
)

type WebserviceHandler struct {
	Logger lib.Logger

	store *sessions.CookieStore
}

func NewWebserverHandler(logger lib.Logger) *WebserviceHandler {
	store := sessions.NewCookieStore([]byte(sessionKey))
	gob.Register(&lib.User{})

	return &WebserviceHandler{
		Logger: logger,
		store:  store,
	}
}

func (handler *WebserviceHandler) Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (handler *WebserviceHandler) Log(msg string) {
	handler.Logger.Log(msg)
}

func (handler *WebserviceHandler) Error(w http.ResponseWriter, err error) {
	handler.Logger.Error(err)
	w.WriteHeader(http.StatusInternalServerError)
}

func (handler *WebserviceHandler) StartSession(ctx context.Context, w http.ResponseWriter, r *http.Request, user *lib.User) error {
	session, err := handler.store.Get(r, sessionName)
	if err != nil {
		return err
	}
	session.Values["User"] = user
	session.Save(r, w)

	return nil
}

func (handler *WebserviceHandler) EndSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	session, err := handler.store.Get(r, sessionName)
	if err != nil {
		return err
	}
	session.Values["User"] = nil
	session.Save(r, w)

	return nil
}

func (handler *WebserviceHandler) AddHandler(url string, authenticated bool, fn func(context.Context, http.ResponseWriter, *http.Request)) {

	http.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		// init the main context
		ctx := context.Background()

		if authenticated {
			session, err := handler.store.Get(r, sessionName)
			if err != nil {
				handler.Error(w, err)
				return
			}
			if session.Values["User"] == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			user, ok := session.Values["User"].(*lib.User)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, ctxUser, user)
		}

		fn(ctx, w, r)
	})
}
