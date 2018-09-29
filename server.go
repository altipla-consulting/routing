package routing

import (
	"context"
	"net/http"
	"text/template"

	"github.com/altipla-consulting/sentry"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// Handler should be implemented by the handler functions that we want to register.
type Handler func(w http.ResponseWriter, r *http.Request) error

// ServerOption is implement by any option that can be passed when constructing a new server.
type ServerOption func(server *Server)

// WithBetaAuth installs a rough authentication mechanism to avoid the final users
// to access beta sites.
func WithBetaAuth(username, password string) ServerOption {
	return func(server *Server) {
		server.username = username
		server.password = password
	}
}

// WithLogrus enables logging of the errors of the handlers.
func WithLogrus() ServerOption {
	return func(server *Server) {
		server.logging = true
	}
}

// WithSentry configures Sentry logging of issues in the handlers.
func WithSentry(dsn string) ServerOption {
	return func(server *Server) {
		if dsn != "" {
			server.sentryClient = sentry.NewClient(dsn)
		}
	}
}

// Server configures the routing table.
type Server struct {
	router *httprouter.Router

	// Options
	username, password string
	sentryClient       *sentry.Client
	logging            bool
}

// NewServer configures a new router with the options.
func NewServer(opts ...ServerOption) *Server {
	r := httprouter.New()
	r.NotFound = http.HandlerFunc(NotFoundHandler)

	s := &Server{
		router: r,
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Router returns the raw underlying router to make advanced modifications.
// If you modify the NotFound handler remember to call routing.NotFoundHandler
// as fallback if you don't want to process the request.
func (s *Server) Router() *httprouter.Router {
	return s.router
}

// Get registers a new GET route in the router.
func (s *Server) Get(lang, path string, handler Handler) {
	s.router.GET(path, s.decorate(lang, handler))
}

// Post registers a new POST route in the router.
func (s *Server) Post(lang, path string, handler Handler) {
	s.router.POST(path, s.decorate(lang, handler))
}

// Put registers a new PUT route in the router.
func (s *Server) Put(lang, path string, handler Handler) {
	s.router.PUT(path, s.decorate(lang, handler))
}

// Delete registers a new DELETE route in the router.
func (s *Server) Delete(lang, path string, handler Handler) {
	s.router.DELETE(path, s.decorate(lang, handler))
}

// Group registers all the routes of the group in the router.
func (s *Server) Group(g Group) {
	for lang, url := range g.URL {
		h := func(w http.ResponseWriter, r *http.Request) error {
			r = r.WithContext(context.WithValue(r.Context(), groupKey, g))

			return g.Handler(w, r)
		}

		switch g.Method {
		case http.MethodGet:
			s.Get(lang, url, h)
		case http.MethodPost:
			s.Post(lang, url, h)
		case http.MethodDelete:
			s.Delete(lang, url, h)
		default:
			s.Get(lang, url, h)
		}
	}
}

// NotFound registers a custom NotFound handler in the routing table. Call
// routing.NotFoundHandler as fallback inside your handler if you need it.
func (s *Server) NotFound(lang string, handler Handler) {
	s.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.decorate(lang, handler)(w, r, nil)
	})
}

func (s *Server) decorate(lang string, handler Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		r = r.WithContext(context.WithValue(r.Context(), requestKey, r))
		r = r.WithContext(context.WithValue(r.Context(), paramsKey, ps))
		r = r.WithContext(context.WithValue(r.Context(), langKey, lang))

		if s.username != "" && s.password != "" {
			if _, err := r.Cookie("routing.beta"); err != nil && err != http.ErrNoCookie {
				log.WithField("error", err.Error()).Error("Cannot read cookie")
				emitPage(w, http.StatusInternalServerError)
				return
			} else if err == http.ErrNoCookie {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

				username, password, ok := r.BasicAuth()
				if !ok {
					emitPage(w, http.StatusUnauthorized)
					return
				}
				if username != s.username || password != s.password {
					emitPage(w, http.StatusUnauthorized)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:  "routing.beta",
					Value: "beta",
				})
			}
		}

		if err := handler(w, r); err != nil {
			if s.logging {
				log.WithField("error", err.Error()).Errorf("Handler failed")
			}

			if httperr, ok := err.(Error); ok {
				switch httperr.StatusCode {
				case http.StatusNotFound, http.StatusUnauthorized, http.StatusBadRequest:
					emitPage(w, httperr.StatusCode)
					return
				default:
					err = Internal("unknown routing error: %s", err)
				}
			} else {
				err = Internal("internal error: %s", err)
			}

			if s.sentryClient != nil {
				s.sentryClient.ReportRequest(err, r)
			}

			emitPage(w, http.StatusInternalServerError)
		}
	}
}

func emitPage(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "text/html")
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

// NotFoundHandler is configured inside the router as the default 404 page. If you
// change the handler you can call this function as a fallback.
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	emitPage(w, http.StatusNotFound)
}
