package git

import (
	"bytes"
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

// Repo is a git repository
type Repo struct {
	path string
	Meta RepoMeta
}

// Name returns the human-friendly name, the dir name without the .git suffix
func (r Repo) Name() string {
	return strings.TrimSuffix(filepath.Base(r.path), ".git")
}

// Path returns the path to the Repo
func (r Repo) Path() string {
	return r.path
}

// NewRepo constructs a Repo given a dir and name
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

// Commit is a git commit
type Commit struct {
	SHA       string
	Message   string
	Signature string
	Author    string
	Email     string
	When      time.Time

	// Extra
	Stats CommitStats
	Patch string
	Files []CommitFile
}

// CommitStats is the stats of a commit
type CommitStats struct {
	Changed   int
	Additions int
	Deletions int
}

// CommitFile is a file contained in a commit
type CommitFile struct {
	From   CommitFileEntry
	To     CommitFileEntry
	Action string
	Patch  string
}

// CommitFileEntry is a from/to in a file commit
type CommitFileEntry struct {
	Path   string
	Commit string
}

// Path returns either the To or From path, in order of preference
func (c CommitFile) Path() string {
	if c.To.Path != "" {
		return c.To.Path
	}
	return c.From.Path
}

// Short returns the first eight characters of the SHA
func (c Commit) Short() string {
	return c.SHA[:8]
}

// Summary returns the first line of the commit, suitable for a <summary>
func (c Commit) Summary() string {
	return strings.Split(c.Message, "\n")[0]
}

// Details returns all lines *after* the first, suitable for <details>
func (c Commit) Details() string {
	return strings.Join(strings.Split(c.Message, "\n")[1:], "\n")
}

// Commit gets a specific commit by SHA, including all commit information
func (r Repo) Commit(sha string) (Commit, error) {
	repo, err := r.Git()
	if err != nil {
		return Commit{}, err
	}

	return commit(repo, sha, true)
}

// LastCommit returns the last commit of the repo without any extra information
func (r Repo) LastCommit() (Commit, error) {
	repo, err := r.Git()
	if err != nil {
		return Commit{}, err
	}

	head, err := repo.Head()
	if err != nil {
		return Commit{}, err
	}

	return commit(repo, head.Hash().String(), false)
}

func commit(repo *git.Repository, sha string, extra bool) (Commit, error) {
	obj, err := repo.CommitObject(plumbing.NewHash(sha))
	if err != nil {
		return Commit{}, err
	}

	var c, a, d int
	var p string
	var f []CommitFile
	if extra {
		stats, err := obj.Stats()
		if err != nil {
			return Commit{}, err
		}

		c = len(stats)
		for _, stat := range stats {
			a += stat.Addition
			d += stat.Deletion
		}

		parent, err := obj.Parent(0)
		if err != nil {
			return Commit{}, err
		}

		patch, err := obj.Patch(parent)
		if err != nil {
			return Commit{}, err
		}

		var buf bytes.Buffer
		if err := patch.Encode(&buf); err != nil {
			return Commit{}, err
		}
		p = buf.String()

		objTree, err := obj.Tree()
		if err != nil {
			return Commit{}, err
		}
		parentTree, err := parent.Tree()
		if err != nil {
			return Commit{}, err
		}

		changes, err := parentTree.Diff(objTree)
		if err != nil {
			return Commit{}, err
		}

		for _, change := range changes {
			action, err := change.Action()
			if err != nil {
				return Commit{}, err
			}
			patch, err := change.Patch()
			if err != nil {
				return Commit{}, err
			}
			var buf bytes.Buffer
			if err := patch.Encode(&buf); err != nil {
				return Commit{}, err
			}
			f = append(f, CommitFile{
				From: CommitFileEntry{
					Path:   change.From.Name,
					Commit: parent.Hash.String(),
				},
				To: CommitFileEntry{
					Path:   change.To.Name,
					Commit: obj.Hash.String(),
				},
				Action: action.String(),
				Patch:  buf.String(),
			})
		}
	}

	return Commit{
		SHA:       obj.Hash.String(),
		Message:   obj.Message,
		Signature: obj.PGPSignature,
		Author:    obj.Author.Name,
		Email:     obj.Author.Email,
		When:      obj.Author.When,
		Stats: CommitStats{
			Changed:   c,
			Additions: a,
			Deletions: d,
		},
		Patch: p,
		Files: f,
	}, nil
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
		switch {
		case errors.Is(err, plumbing.ErrObjectNotFound):
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
		case err == nil:
			tags = append(tags, Tag{
				Name:       obj.Name,
				Annotation: obj.Message,
				Signature:  obj.PGPSignature,
				When:       obj.Tagger.When,
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

// Commits returns commits from a specific hash in descending order
func (r Repo) Commits(ref string) ([]Commit, error) {
	repo, err := r.Git()
	if err != nil {
		return nil, err
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, err
	}

	cmts, err := repo.Log(&git.LogOptions{
		From: *hash,
	})
	if err != nil {
		return nil, err
	}

	var commits []Commit
	if err := cmts.ForEach(func(commit *object.Commit) error {
		commits = append(commits, Commit{
			SHA:       commit.Hash.String(),
			Message:   commit.Message,
			Signature: commit.PGPSignature,
			Author:    commit.Author.Name,
			Email:     commit.Author.Email,
			When:      commit.Author.When,
		})
		return nil
	}); err != nil {
		return nil, err
	}

	return commits, nil
}
