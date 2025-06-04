package git_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
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

func TestPathExists(t *testing.T) {
	tmp := t.TempDir()
	exists, err := git.PathExists(tmp)
	assert.NoError(t, err)
	assert.True(t, exists)

	doesNotExist := filepath.Join(tmp, "does-not-exist")
	exists, err = git.PathExists(doesNotExist)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRepoMetaUpdate(t *testing.T) {
	original := git.RepoMeta{
		Description: "Original description",
		Private:     true,
		Tags:        git.TagSet{"tag1": struct{}{}, "tag2": struct{}{}},
	}

	update := git.RepoMeta{
		Description: "Updated description",
		Private:     false,
		Tags:        git.TagSet{"tag3": struct{}{}},
	}

	err := original.Update(update)
	assert.NoError(t, err)

	assert.Equal(t, "Updated description", original.Description)
	assert.False(t, original.Private)
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, original.Tags.Slice())
}

func TestFileInfoName(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{path: "file.txt", expected: "file.txt"},
		{path: "dir/file.txt", expected: "file.txt"},
		{path: "nested/path/to/file.go", expected: "file.go"},
		{path: "README.md", expected: "README.md"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			fi := git.FileInfo{Path: tc.path}
			assert.Equal(t, tc.expected, fi.Name())
		})
	}
}

func TestCommitSummaryAndDetails(t *testing.T) {
	testCases := []struct {
		message         string
		expectedSummary string
		expectedDetails string
	}{
		{
			message:         "Simple commit message",
			expectedSummary: "Simple commit message",
			expectedDetails: "",
		},
		{
			message:         "Add feature X\n\nThis commit adds feature X\nWith multiple details\nAcross multiple lines",
			expectedSummary: "Add feature X",
			expectedDetails: "\nThis commit adds feature X\nWith multiple details\nAcross multiple lines",
		},
		{
			message:         "Fix bug\n\nDetailed explanation",
			expectedSummary: "Fix bug",
			expectedDetails: "\nDetailed explanation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			commit := git.Commit{
				SHA:       "abcdef1234567890",
				Message:   tc.message,
				Signature: "",
				Author:    "Test User",
				Email:     "test@example.com",
				When:      time.Now(),
			}

			assert.Equal(t, tc.expectedSummary, commit.Summary())
			assert.Equal(t, tc.expectedDetails, commit.Details())
		})
	}
}

func TestCommitShort(t *testing.T) {
	commit := git.Commit{
		SHA: "abcdef1234567890abcdef1234567890",
	}

	assert.Equal(t, "abcdef12", commit.Short())
}

func TestCommitFilePath(t *testing.T) {
	testCases := []struct {
		name     string
		fromPath string
		toPath   string
		expected string
	}{
		{
			name:     "to path preferred",
			fromPath: "old/path.txt",
			toPath:   "new/path.txt",
			expected: "new/path.txt",
		},
		{
			name:     "fallback to from path",
			fromPath: "deleted/file.txt",
			toPath:   "",
			expected: "deleted/file.txt",
		},
		{
			name:     "both paths empty",
			fromPath: "",
			toPath:   "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cf := git.CommitFile{
				From: git.CommitFileEntry{Path: tc.fromPath},
				To:   git.CommitFileEntry{Path: tc.toPath},
			}
			assert.Equal(t, tc.expected, cf.Path())
		})
	}
}

func TestRepoName(t *testing.T) {
	tmp := t.TempDir()

	repoName := "testrepo"
	err := git.EnsureRepo(tmp, repoName+".git")
	assert.NoError(t, err)

	repo, err := git.NewRepo(tmp, repoName)
	assert.NoError(t, err)
	assert.Equal(t, repoName, repo.Name())

	repoName2 := "test-repo-with-hyphens"
	err = git.EnsureRepo(tmp, repoName2+".git")
	assert.NoError(t, err)

	repo2, err := git.NewRepo(tmp, repoName2)
	assert.NoError(t, err)
	assert.Equal(t, repoName2, repo2.Name())
}

func TestHandlePushOptions(t *testing.T) {
	tmp := t.TempDir()
	err := git.EnsureRepo(tmp, "test.git")
	assert.NoError(t, err)

	repo, err := git.NewRepo(tmp, "test")
	assert.NoError(t, err)

	opts := []*packp.Option{
		{Key: "description", Value: "New description"},
	}
	err = git.HandlePushOptions(repo, opts)
	assert.NoError(t, err)
	assert.Equal(t, "New description", repo.Meta.Description)

	opts = []*packp.Option{
		{Key: "private", Value: "false"},
	}
	err = git.HandlePushOptions(repo, opts)
	assert.NoError(t, err)
	assert.False(t, repo.Meta.Private)

	repo.Meta.Private = true
	opts = []*packp.Option{
		{Key: "private", Value: "invalid"},
	}
	err = git.HandlePushOptions(repo, opts)
	assert.NoError(t, err)
	assert.True(t, repo.Meta.Private)

	opts = []*packp.Option{
		{Key: "tags", Value: "tag1,tag2"},
	}
	err = git.HandlePushOptions(repo, opts)
	assert.NoError(t, err)

	opts = []*packp.Option{
		{Key: "description", Value: "Combined update"},
		{Key: "private", Value: "true"},
	}
	err = git.HandlePushOptions(repo, opts)
	assert.NoError(t, err)
	assert.Equal(t, "Combined update", repo.Meta.Description)
	assert.True(t, repo.Meta.Private)
}

func TestRepoPath(t *testing.T) {
	tmp := t.TempDir()
	err := git.EnsureRepo(tmp, "test.git")
	assert.NoError(t, err)

	repo, err := git.NewRepo(tmp, "test")
	assert.NoError(t, err)

	expected := filepath.Join(tmp, "test.git")
	assert.Equal(t, expected, repo.Path())
}

func TestEnsureJSONFile(t *testing.T) {
	tmp := t.TempDir()
	err := git.EnsureRepo(tmp, "test.git")
	assert.NoError(t, err)

	repo, err := git.NewRepo(tmp, "test")
	assert.NoError(t, err)

	assert.True(t, repo.Meta.Private, "default repo should be private")
	assert.Equal(t, "", repo.Meta.Description, "default description should be empty")
	assert.Equal(t, 0, len(repo.Meta.Tags), "default tags should be empty")
}
