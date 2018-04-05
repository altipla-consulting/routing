package routing

import (
	"net/http"
	"strings"

	"github.com/juju/errors"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

type Group struct {
	Handler Handler
	URL     map[string]string
}

func (g Group) Register(r *httprouter.Router) {
	for lang, url := range g.URL {
		r.GET(url, Middlewares(lang, func(w http.ResponseWriter, r *http.Request) error {
			r = r.WithContext(context.WithValue(r.Context(), groupKey, g))

			return errors.Trace(g.Handler(w, r))
		}))
	}
}

func (g Group) ResolveURL(r *http.Request, lang string) string {
	segments := strings.Split(g.URL[lang], "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[i] = Param(r, segment[1:])
		}
	}

	return strings.Join(segments, "/")
}
