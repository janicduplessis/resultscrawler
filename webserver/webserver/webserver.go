package webserver

import (
	"encoding/gob"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"code.google.com/p/go.net/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/janicduplessis/resultscrawler/lib"
)

const (
	sessionKey  = "AWDoijwapajwdvz23423oaawc"
	sessionName = "ct-session"
	ctxUser     = "User"
)

// Webserver handles all web requests interaction.
type Webserver struct {
	Logger    lib.Logger
	Router    *mux.Router
	Templates *template.Template

	store *sessions.CookieStore
}

// NewWebserver creates a new webserver handler.
func NewWebserver(logger lib.Logger) *Webserver {
	store := sessions.NewCookieStore([]byte(sessionKey))
	gob.Register(&lib.User{})

	router := mux.NewRouter()
	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// Register router
	http.Handle("/", router)

	templates := template.New("template")

	err := filepath.Walk("views", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".html") {
			_, err = templates.ParseFiles(path)
		}

		return err
	})
	if err != nil {
		panic(err)
	}

	return &Webserver{
		Logger:    logger,
		Router:    router,
		Templates: templates,
		store:     store,
	}
}

// Redirect redirects the user to the specified url.
func (handler *Webserver) Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Log logs a message to the main logger.
func (handler *Webserver) Log(msg string) {
	handler.Logger.Log(msg)
}

// Error logs an error and writes an internal server error response.
func (handler *Webserver) Error(w http.ResponseWriter, err error) {
	handler.Logger.Error(err)
	w.WriteHeader(http.StatusInternalServerError)
}

// StartSession logs in the user.
func (handler *Webserver) StartSession(ctx context.Context, w http.ResponseWriter, r *http.Request, user *lib.User) error {
	session, err := handler.store.Get(r, sessionName)
	if err != nil {
		return err
	}
	session.Values["User"] = user
	session.Save(r, w)

	return nil
}

// EndSession logs out the user.
func (handler *Webserver) EndSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	session, err := handler.store.Get(r, sessionName)
	if err != nil {
		return err
	}
	session.Values["User"] = nil
	session.Save(r, w)

	return nil
}

// MakeHandler creates a request handler function that provides a context and
// can be authenticated.
func (handler *Webserver) MakeHandler(authenticated bool, fn func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}
