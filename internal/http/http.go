package http

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.jolheiser.com/ugit/assets"
	"go.jolheiser.com/ugit/internal/git"
	"go.jolheiser.com/ugit/internal/html"
	"go.jolheiser.com/ugit/internal/http/httperr"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server is the container struct for the HTTP server
type Server struct {
	port int
	mux  *chi.Mux
}

// ListenAndServe simply wraps http.ListenAndServe to contain the functionality here
func (s Server) ListenAndServe() error {
	return http.ListenAndServe(fmt.Sprintf("localhost:%d", s.port), s.mux)
}

// Settings is the configuration for the HTTP server
type Settings struct {
	Title       string
	Description string
	CloneURL    string
	Port        int
	RepoDir     string
	Profile     Profile
}

// Profile is the index profile
type Profile struct {
	Username string
	Email    string
	Links    []Link
}

// Link is a profile link
type Link struct {
	Name string
	URL  string
}

func (s Settings) goGet(repo string) string {
	u, _ := url.Parse(s.CloneURL)
	return fmt.Sprintf(`<!DOCTYPE html><title>%[1]s</title><meta name="go-import" content="%[2]s/%[1]s git %[3]s/%[1]s.git"><meta name="go-source" content="%[2]s/%[1]s _ %[3]s/%[1]s/tree/main{/dir}/{file}#L{line}">`, repo, u.Hostname(), s.CloneURL)
}

// New returns a new HTTP server
func New(settings Settings) Server {
	mux := chi.NewMux()

	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	rh := repoHandler{s: settings}
	mux.Route("/", func(r chi.Router) {
		r.Get("/", httperr.Handler(rh.index))
		r.Route("/{repo}", func(r chi.Router) {
			r.Use(rh.repoMiddleware)
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				repo := r.Context().Value(repoCtxKey).(*git.Repo)
				if r.URL.Query().Has("go-get") {
					w.Write([]byte(settings.goGet(repo.Name())))
					return
				}
				if strings.HasSuffix(chi.URLParam(r, "repo"), ".git") {
					http.Redirect(w, r, "/"+repo.Name(), http.StatusFound)
					return
				}
				rh.repoTree("", "").ServeHTTP(w, r)
			})
			r.Get("/tree/{ref}/*", func(w http.ResponseWriter, r *http.Request) {
				rh.repoTree(chi.URLParam(r, "ref"), chi.URLParam(r, "*")).ServeHTTP(w, r)
			})
			r.Get("/refs", httperr.Handler(rh.repoRefs))
			r.Get("/log/{ref}", httperr.Handler(rh.repoLog))
			r.Get("/commit/{commit}", httperr.Handler(rh.repoCommit))
			r.Get("/commit/{commit}.patch", httperr.Handler(rh.repoPatch))
			r.Get("/search", httperr.Handler(rh.repoSearch))

			// Protocol
			r.Get("/info/refs", httperr.Handler(rh.infoRefs))
			r.Post("/git-upload-pack", httperr.Handler(rh.uploadPack))
		})
	})

	mux.Route("/_", func(r chi.Router) {
		r.Get("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Write(assets.LogoIcon)
		})
		r.Get("/tailwind.css", html.TailwindHandler)
	})

	return Server{mux: mux, port: settings.Port}
}

type repoHandler struct {
	s Settings
}

func (rh repoHandler) baseContext() html.BaseContext {
	return html.BaseContext{
		Title:       rh.s.Title,
		Description: rh.s.Description,
	}
}

func (rh repoHandler) repoHeaderContext(repo *git.Repo, r *http.Request) html.RepoHeaderComponentContext {
	ref := chi.URLParam(r, "ref")
	if ref == "" {
		ref, _ = repo.DefaultBranch()
	}
	return html.RepoHeaderComponentContext{
		Description: repo.Meta.Description,
		Name:        chi.URLParam(r, "repo"),
		Ref:         ref,
		CloneURL:    rh.s.CloneURL,
	}
}

func (rh repoHandler) repoBreadcrumbContext(repo *git.Repo, r *http.Request, path string) html.RepoBreadcrumbComponentContext {
	ref := chi.URLParam(r, "ref")
	if ref == "" {
		ref, _ = repo.DefaultBranch()
	}
	return html.RepoBreadcrumbComponentContext{
		Repo: chi.URLParam(r, "repo"),
		Ref:  ref,
		Path: path,
	}
}

// NoopLogger is a no-op logging middleware
func NoopLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
