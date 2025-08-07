package http

import (
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"go.jolheiser.com/ugit/internal/git"
	"go.jolheiser.com/ugit/internal/html"
	"go.jolheiser.com/ugit/internal/http/httperr"
)

func (rh repoHandler) index(w http.ResponseWriter, r *http.Request) error {
	repoPaths, err := os.ReadDir(rh.s.RepoDir)
	if err != nil {
		return httperr.Error(err)
	}

	tagFilter := r.URL.Query().Get("tag")

	repos := make([]*git.Repo, 0, len(repoPaths))
	for _, repoName := range repoPaths {
		if !strings.HasSuffix(repoName.Name(), ".git") {
			continue
		}
		repo, err := git.NewRepo(rh.s.RepoDir, repoName.Name())
		if err != nil {
			return httperr.Error(err)
		}
		if repo.Meta.Private {
			if !rh.s.ShowPrivate {
				continue
			}
			repo.Meta.Tags.Add("private")
		}

		if tagFilter != "" && !repo.Meta.Tags.Contains(strings.ToLower(tagFilter)) {
			continue
		}
		repos = append(repos, repo)
	}
	sort.Slice(repos, func(i, j int) bool {
		var when1, when2 time.Time
		if c, err := repos[i].LastCommit(); err == nil {
			when1 = c.When
		}
		if c, err := repos[j].LastCommit(); err == nil {
			when2 = c.When
		}
		return when1.After(when2)
	})

	links := make([]html.IndexLink, 0, len(rh.s.Profile.Links))
	for _, link := range rh.s.Profile.Links {
		links = append(links, html.IndexLink{
			Name: link.Name,
			URL:  link.URL,
		})
	}

	if err := html.Index(html.IndexContext{
		BaseContext: rh.baseContext(),
		Profile: html.IndexProfile{
			Username: rh.s.Profile.Username,
			Email:    rh.s.Profile.Email,
			Links:    links,
		},
		Repos: repos,
	}).Render(r.Context(), w); err != nil {
		return httperr.Error(err)
	}

	return nil
}
