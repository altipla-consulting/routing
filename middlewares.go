package routing

import (
	"net/http"
	"os"
	"text/template"

	"github.com/juju/errors"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func Middlewares(lang string, handler Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		r = r.WithContext(context.WithValue(r.Context(), requestKey, r))
		r = r.WithContext(context.WithValue(r.Context(), paramsKey, ps))
		r = r.WithContext(context.WithValue(r.Context(), langKey, lang))

		if os.Getenv("VERSION") == "beta" {
			ok, err := checkAuth(w, r)
			if err != nil {
				emitPage(w, http.StatusInternalServerError)
				return
			}
			if !ok {
				return
			}
		}

		if err := handler(w, r); err != nil {
			if errors.IsUnauthorized(err) {
				log.WithField("error", err.Error()).Error("Unauthorized")
				emitPage(w, http.StatusUnauthorized)
				return
			}
			if errors.IsNotValid(err) {
				log.WithField("error", err.Error()).Error("Bad Request")
				emitPage(w, http.StatusBadRequest)
				return
			}
			if errors.IsNotFound(err) {
				log.WithField("error", err.Error()).Error("Not Found")
				emitPage(w, http.StatusNotFound)
				return
			}

			log.WithField("error", err.Error()).Error("Internal Server Error")
			emitPage(w, http.StatusInternalServerError)
			return
		}
	}
}

func checkAuth(w http.ResponseWriter, r *http.Request) (bool, error) {
	if _, err := r.Cookie("staging.auth"); err != nil && err != http.ErrNoCookie {
		return false, errors.Trace(err)
	}

	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	username, password, ok := r.BasicAuth()
	if !ok {
		emitPage(w, http.StatusUnauthorized)
		return false, nil
	}

	if username != "adoquier" || password != "webs" {
		emitPage(w, http.StatusUnauthorized)
		return false, nil
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "staging.auth",
		Value: "beta",
	})
	return true, nil
}

func emitPage(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	tmpl, err := template.New("error").Parse(errorTemplate)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.WithField("error", err.Error()).Error("Cannot parse template")
	}
	if err := tmpl.Execute(w, status); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.WithField("error", err.Error()).Error("Cannot execute template")
	}
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	emitPage(w, http.StatusNotFound)
}
