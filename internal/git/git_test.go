package git_test

import (
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.jolheiser.com/ugit/internal/git"
)

func TestEnsureRepo(t *testing.T) {
	tmp := t.TempDir()

	ok, err := git.PathExists(filepath.Join(tmp, "test"))
	assert.False(t, ok, "repo should not exist yet")
	assert.NoError(t, err, "PathExists should not error when repo doesn't exist")

	err = git.EnsureRepo(tmp, "test")
	assert.NoError(t, err, "repo should be created")

	ok, err = git.PathExists(filepath.Join(tmp, "test"))
	assert.True(t, ok, "repo should exist")
	assert.NoError(t, err, "EnsureRepo should not error when path exists")

	err = git.EnsureRepo(tmp, "test")
	assert.NoError(t, err, "repo should already exist")
}

func TestRepo(t *testing.T) {
	tmp := t.TempDir()
	err := git.EnsureRepo(tmp, "test.git")
	assert.NoError(t, err, "should create repo")

	repo, err := git.NewRepo(tmp, "test")
	assert.NoError(t, err, "should init new repo")
	assert.True(t, repo.Meta.Private, "repo should default to private")

	repo.Meta.Private = false
	err = repo.SaveMeta()
	assert.NoError(t, err, "should save repo meta")

	repo, err = git.NewRepo(tmp, "test")
	assert.NoError(t, err, "should not error when getting existing repo")
	assert.False(t, repo.Meta.Private, "repo should be public after saving meta")
}
