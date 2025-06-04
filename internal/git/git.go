package git

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// EnsureRepo ensures that the repo exists in the given directory
func EnsureRepo(dir string, repo string) error {
	exists, err := PathExists(dir)
	if err != nil {
		return err
	}
	if !exists {
		err = os.MkdirAll(dir, os.ModeDir|os.FileMode(0o700))
		if err != nil {
			return err
		}
	}
	rp := filepath.Join(dir, repo)
	exists, err = PathExists(rp)
	if err != nil {
		return err
	}
	if !exists {
		_, err := git.PlainInit(rp, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// PathExists checks if a path exists and returns true if it does
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return true, err
}

// Tree returns the git tree at a given ref/rev
func (r Repo) Tree(ref string) (*object.Tree, error) {
	g, err := r.Git()
	if err != nil {
		return nil, err
	}

	hash, err := g.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, err
	}

	c, err := g.CommitObject(*hash)
	if err != nil {
		return nil, err
	}

	return c.Tree()
}

// FileInfo is the information for a file in a tree
type FileInfo struct {
	Path  string
	IsDir bool
	Mode  string
	Size  string
}

// Name returns the last part of the FileInfo.Path
func (f FileInfo) Name() string {
	return filepath.Base(f.Path)
}

// Dir returns the given dirpath in the given ref as a slice of FileInfo
// Sorted alphabetically, dirs first
func (r Repo) Dir(ref, path string) ([]FileInfo, error) {
	t, err := r.Tree(ref)
	if err != nil {
		return nil, err
	}
	if path != "" {
		t, err = t.Tree(path)
		if err != nil {
			return nil, err
		}
	}

	fis := make([]FileInfo, 0, len(t.Entries))
	for _, entry := range t.Entries {
		fm, err := entry.Mode.ToOSFileMode()
		if err != nil {
			return nil, err
		}
		size, err := t.Size(entry.Name)
		if err != nil {
			return nil, err
		}
		fis = append(fis, FileInfo{
			Path:  filepath.Join(path, entry.Name),
			IsDir: fm.IsDir(),
			Mode:  fm.String(),
			Size:  humanize.Bytes(uint64(size)),
		})
	}
	sort.Slice(fis, func(i, j int) bool {
		fi1 := fis[i]
		fi2 := fis[j]
		if fi1.IsDir != fi2.IsDir {
			return fi1.IsDir
		}
		return fi1.Name() < fi2.Name()
	})

	return fis, nil
}

// GetCommitFromRef returns the commit object for a given ref
func (r Repo) GetCommitFromRef(ref string) (*object.Commit, error) {
	g, err := r.Git()
	if err != nil {
		return nil, err
	}

	hash, err := g.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, err
	}

	return g.CommitObject(*hash)
}

// FileContent returns the content of a file in the git tree at a given ref/rev
func (r Repo) FileContent(ref, file string) (string, error) {
	t, err := r.Tree(ref)
	if err != nil {
		return "", err
	}

	f, err := t.File(file)
	if err != nil {
		return "", err
	}

	content, err := f.Contents()
	if err != nil {
		return "", err
	}

	return content, nil
}
