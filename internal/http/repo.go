package http

import (
	"bytes"
	"errors"
	"go.jolheiser.com/ugit/internal/html/markup"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"

	"go.jolheiser.com/ugit/internal/git"
	"go.jolheiser.com/ugit/internal/html"
	"go.jolheiser.com/ugit/internal/http/httperr"

	"github.com/go-chi/chi/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func (rh repoHandler) repoTree(ref, path string) http.HandlerFunc {
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

		if ref == "" {
			ref, err = repo.DefaultBranch()
			if err != nil {
				return httperr.Error(err)
			}
		}

		tree, err := repo.Dir(ref, path)
		if err != nil {
			if errors.Is(err, object.ErrDirectoryNotFound) {
				return rh.repoFile(w, r, repo, ref, path)
			}
			return httperr.Error(err)
		}

		readmeContent, err := markup.Readme(repo, ref, path)
		if err != nil {
			return httperr.Error(err)
		}

		var back string
		if path != "" {
			back = filepath.Dir(path)
		}
		if err := html.RepoTree(html.RepoTreeContext{
			Description:                repo.Meta.Description,
			BaseContext:                rh.baseContext(),
			RepoHeaderComponentContext: rh.repoHeaderContext(repo, r),
			RepoTreeComponentContext: html.RepoTreeComponentContext{
				Repo: repoName,
				Ref:  ref,
				Tree: tree,
				Back: back,
			},
			ReadmeComponentContext: html.ReadmeComponentContext{
				Markdown: readmeContent,
			},
		}).Render(r.Context(), w); err != nil {
			return httperr.Error(err)
		}

		return nil
	})
}

func (rh repoHandler) repoFile(w http.ResponseWriter, r *http.Request, repo *git.Repo, ref, path string) error {
	content, err := repo.FileContent(ref, path)
	if err != nil {
		return httperr.Error(err)
	}

	if r.URL.Query().Has("raw") {
		if r.URL.Query().Has("pretty") {
			ext := filepath.Ext(path)
			w.Header().Set("Content-Type", mime.TypeByExtension(ext))
		}
		w.Write([]byte(content))
		return nil
	}

	var buf bytes.Buffer
	if err := markup.Code.Convert([]byte(content), filepath.Base(path), &buf); err != nil {
		return httperr.Error(err)
	}

	if err := html.RepoFile(html.RepoFileContext{
		BaseContext:                rh.baseContext(),
		RepoHeaderComponentContext: rh.repoHeaderContext(repo, r),
		Code:                       buf.String(),
		Path:                       path,
	}).Render(r.Context(), w); err != nil {
		return httperr.Error(err)
	}

	return nil
}

func (rh repoHandler) repoRefs(w http.ResponseWriter, r *http.Request) error {
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

	branches, err := repo.Branches()
	if err != nil {
		return httperr.Error(err)
	}

	tags, err := repo.Tags()
	if err != nil {
		return httperr.Error(err)
	}

	if err := html.RepoRefs(html.RepoRefsContext{
		BaseContext:                rh.baseContext(),
		RepoHeaderComponentContext: rh.repoHeaderContext(repo, r),
		Branches:                   branches,
		Tags:                       tags,
	}).Render(r.Context(), w); err != nil {
		return httperr.Error(err)
	}

	return nil
}

func (rh repoHandler) repoLog(w http.ResponseWriter, r *http.Request) error {
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

	commits, err := repo.Commits(chi.URLParam(r, "ref"))
	if err != nil {
		return httperr.Error(err)
	}

	if err := html.RepoLog(html.RepoLogContext{
		BaseContext:                rh.baseContext(),
		RepoHeaderComponentContext: rh.repoHeaderContext(repo, r),
		Commits:                    commits,
	}).Render(r.Context(), w); err != nil {
		return httperr.Error(err)
	}

	return nil
}

func (rh repoHandler) repoCommit(w http.ResponseWriter, r *http.Request) error {
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

	commit, err := repo.Commit(chi.URLParam(r, "commit"))
	if err != nil {
		return httperr.Error(err)
	}

	for idx, p := range commit.Files {
		var patch bytes.Buffer
		if err := markup.Code.Basic([]byte(p.Patch), "commit.patch", &patch); err != nil {
			return httperr.Error(err)
		}
		commit.Files[idx].Patch = patch.String()
	}

	if err := html.RepoCommit(html.RepoCommitContext{
		BaseContext:                rh.baseContext(),
		RepoHeaderComponentContext: rh.repoHeaderContext(repo, r),
		Commit:                     commit,
	}).Render(r.Context(), w); err != nil {
		return httperr.Error(err)
	}

	return nil
}

func (rh repoHandler) repoPatch(w http.ResponseWriter, r *http.Request) error {
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

	commit, err := repo.Commit(chi.URLParam(r, "commit"))
	if err != nil {
		return httperr.Error(err)
	}

	w.Write([]byte(commit.Patch))

	return nil
}
