package http

import (
	"context"
	"errors"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.jolheiser.com/ugit/internal/git"
	"go.jolheiser.com/ugit/internal/http/httperr"
)

type ugitCtxKey string

var repoCtxKey = ugitCtxKey("repo")

func (rh repoHandler) repoMiddleware(next http.Handler) http.Handler {
	return httperr.Handler(func(w http.ResponseWriter, r *http.Request) error {
		repoName := chi.URLParam(r, "repo")
		repo, err := git.NewRepo(rh.s.RepoDir, repoName)
		if err != nil {
			httpErr := http.StatusInternalServerError
			if errors.Is(err, fs.ErrNotExist) {
				httpErr = http.StatusNotFound
			}
			return httperr.Status(err, httpErr)
		}
		if repo.Meta.Private {
			return httperr.Status(errors.New("could not get git repo"), http.StatusNotFound)
		}
		r = r.WithContext(context.WithValue(r.Context(), repoCtxKey, repo))
		next.ServeHTTP(w, r)
		return nil
	})
}
