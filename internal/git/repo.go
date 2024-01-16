package git

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repo struct {
	path string
	Meta RepoMeta
}

func (r Repo) Name() string {
	return strings.TrimSuffix(filepath.Base(r.path), ".git")
}

func NewRepo(dir, name string) (*Repo, error) {
	if !strings.HasSuffix(name, ".git") {
		name += ".git"
	}
	r := &Repo{
		path: filepath.Join(dir, name),
	}

	_, err := os.Stat(r.path)
	if err != nil {
		return nil, err
	}

	if err := ensureJSONFile(r.metaPath()); err != nil {
		return nil, err
	}
	fi, err := os.Open(r.metaPath())
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	if err := json.NewDecoder(fi).Decode(&r.Meta); err != nil {
		return nil, err
	}

	return r, nil
}

// DefaultBranch returns the branch referenced by HEAD, setting it if needed
func (r Repo) DefaultBranch() (string, error) {
	repo, err := r.Git()
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		if !errors.Is(err, plumbing.ErrReferenceNotFound) {
			return "", err
		}
		brs, err := repo.Branches()
		if err != nil {
			return "", err
		}
		defer brs.Close()
		fb, err := brs.Next()
		if err != nil {
			return "", err
		}
		// Rename the default branch to the first branch available
		ref = fb
		sym := plumbing.NewSymbolicReference(plumbing.HEAD, fb.Name())
		if err := repo.Storer.SetReference(sym); err != nil {
			return "", err
		}
	}

	return strings.TrimPrefix(ref.Name().String(), "refs/heads/"), nil
}

// Git allows access to the git repository
func (r Repo) Git() (*git.Repository, error) {
	return git.PlainOpen(r.path)
}

// LastCommit returns the last commit of the repo
func (r Repo) LastCommit() (*object.Commit, error) {
	repo, err := r.Git()
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	return repo.CommitObject(head.Hash())
}

// Branches is all repo branches, default first and sorted alphabetically after that
func (r Repo) Branches() ([]string, error) {
	repo, err := r.Git()
	if err != nil {
		return nil, err
	}

	def, err := r.DefaultBranch()
	if err != nil {
		return nil, err
	}

	brs, err := repo.Branches()
	if err != nil {
		return nil, err
	}

	var branches []string
	if err := brs.ForEach(func(branch *plumbing.Reference) error {
		branches = append(branches, branch.Name().Short())
		return nil
	}); err != nil {
		return nil, err
	}

	sort.Slice(branches, func(i, j int) bool {
		return branches[i] == def || branches[i] < branches[j]
	})

	return branches, nil
}

// Tag is a git tag, which may or may not have an annotation/signature
type Tag struct {
	Name       string
	Annotation string
	Signature  string
	When       time.Time
}

// Tags is all repo tags, sorted by time descending
func (r Repo) Tags() ([]Tag, error) {
	repo, err := r.Git()
	if err != nil {
		return nil, err
	}

	tgs, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	var tags []Tag
	if err := tgs.ForEach(func(tag *plumbing.Reference) error {
		obj, err := repo.TagObject(tag.Hash())
		switch err {
		case nil:
			tags = append(tags, Tag{
				Name:       obj.Name,
				Annotation: obj.Message,
				Signature:  obj.PGPSignature,
				When:       obj.Tagger.When,
			})
		case plumbing.ErrObjectNotFound:
			commit, err := repo.CommitObject(tag.Hash())
			if err != nil {
				return err
			}
			tags = append(tags, Tag{
				Name:       tag.Name().Short(),
				Annotation: commit.Message,
				Signature:  commit.PGPSignature,
				When:       commit.Author.When,
			})
		default:
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].When.After(tags[j].When)
	})

	return tags, nil
}
