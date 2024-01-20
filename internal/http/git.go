package http

import (
	"errors"
	"net/http"

	"go.jolheiser.com/ugit/internal/git"
	"go.jolheiser.com/ugit/internal/http/httperr"
)

func (rh repoHandler) infoRefs(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Query().Get("service") != "git-upload-pack" {
		return httperr.Status(errors.New("pushing isn't supported via HTTP(S), use SSH"), http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
	repo := r.Context().Value(repoCtxKey).(*git.Repo)
	protocol, err := git.NewProtocol(repo.Path())
	if err != nil {
		return httperr.Error(err)
	}
	if err := protocol.HTTPInfoRefs(Session{
		w: w,
		r: r,
	}); err != nil {
		return httperr.Error(err)
	}

	return nil
}

func (rh repoHandler) uploadPack(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("content-type", "application/x-git-upload-pack-result")
	repo := r.Context().Value(repoCtxKey).(*git.Repo)
	protocol, err := git.NewProtocol(repo.Path())
	if err != nil {
		return httperr.Error(err)
	}
	if err := protocol.HTTPUploadPack(Session{
		w: w,
		r: r,
	}); err != nil {
		return httperr.Error(err)
	}

	return nil
}
