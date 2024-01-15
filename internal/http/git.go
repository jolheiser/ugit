package http

import (
	"errors"
	"net/http"
	"path/filepath"

	"go.jolheiser.com/ugit/internal/git"
	"go.jolheiser.com/ugit/internal/http/httperr"

	"github.com/go-chi/chi/v5"
)

func (rh repoHandler) infoRefs(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Query().Get("service") != "git-upload-pack" {
		return httperr.Status(errors.New("pushing isn't supported via HTTP(S), use SSH"), http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
	rp := filepath.Join(rh.s.RepoDir, chi.URLParam(r, "repo")+".git")
	repo, err := git.NewProtocol(rp)
	if err != nil {
		return httperr.Error(err)
	}
	if err := repo.HTTPInfoRefs(Session{
		w: w,
		r: r,
	}); err != nil {
		return httperr.Error(err)
	}

	return nil
}

func (rh repoHandler) uploadPack(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("content-type", "application/x-git-upload-pack-result")
	rp := filepath.Join(rh.s.RepoDir, chi.URLParam(r, "repo")+".git")
	repo, err := git.NewProtocol(rp)
	if err != nil {
		return httperr.Error(err)
	}
	if err := repo.HTTPUploadPack(Session{
		w: w,
		r: r,
	}); err != nil {
		return httperr.Error(err)
	}

	return nil
}
