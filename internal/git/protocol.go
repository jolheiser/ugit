package git

import (
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/serverinfo"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

// ReadWriteContexter is the interface required to operate on git protocols
type ReadWriteContexter interface {
	io.ReadWriteCloser
	Context() context.Context
}

type Protocoler interface {
	HTTPInfoRefs(ReadWriteContexter) error
	HTTPUploadPack(ReadWriteContexter) error
	SSHUploadPack(ReadWriteContexter) error
	SSHReceivePack(ReadWriteContexter, *Repo) error
}

// UpdateServerInfo handles updating server info for the git repo
func UpdateServerInfo(repo string) error {
	r, err := git.PlainOpen(repo)
	if err != nil {
		return err
	}
	fs := r.Storer.(*filesystem.Storage).Filesystem()
	return serverinfo.UpdateServerInfo(r.Storer, fs)
}

// HandlePushOptions handles all relevant push options for a [Repo] and saves the new [RepoMeta]
func HandlePushOptions(repo *Repo, opts []*packp.Option) error {
	var changed bool
	for _, opt := range opts {
		switch strings.ToLower(opt.Key) {
		case "desc", "description":
			changed = repo.Meta.Description != opt.Value
			repo.Meta.Description = opt.Value
		case "private":
			private, err := strconv.ParseBool(opt.Value)
			if err != nil {
				continue
			}
			changed = repo.Meta.Private != private
			repo.Meta.Private = private
		case "tags":
			tagValues := strings.Split(opt.Value, ",")
			for _, tagValue := range tagValues {
				var remove bool
				if strings.HasPrefix(tagValue, "-") {
					remove = true
					tagValue = strings.TrimPrefix(tagValue, "-")
				}
				tagValue = strings.ToLower(tagValue)
				if remove {
					repo.Meta.Tags.Remove(tagValue)
				} else {
					repo.Meta.Tags.Add(tagValue)
				}
			}
		}
	}
	if changed {
		return repo.SaveMeta()
	}
	return nil
}
