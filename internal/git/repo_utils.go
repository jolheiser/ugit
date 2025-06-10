package git

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ListRepos returns all directory entries in the given directory
func ListRepos(dir string) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []fs.DirEntry{}, nil
		}
		return nil, err
	}
	return entries, nil
}

// DeleteRepo deletes a git repository from the filesystem
func DeleteRepo(repoPath string) error {
	return os.RemoveAll(repoPath)
}

// RenameRepo renames a git repository
func RenameRepo(repoDir, oldName, newName string) error {
	if !filepath.IsAbs(repoDir) {
		return errors.New("repository directory must be an absolute path")
	}

	if !filepath.IsAbs(oldName) && !filepath.IsAbs(newName) {
		oldPath := filepath.Join(repoDir, oldName)
		if !strings.HasSuffix(oldPath, ".git") {
			oldPath += ".git"
		}

		newPath := filepath.Join(repoDir, newName)
		if !strings.HasSuffix(newPath, ".git") {
			newPath += ".git"
		}

		return os.Rename(oldPath, newPath)
	}

	return errors.New("repository names should not be absolute paths")
}

// RepoPathExists checks if a path exists
func RepoPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}