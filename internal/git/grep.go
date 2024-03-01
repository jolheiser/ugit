package git

import (
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GrepResult is the result of a search
type GrepResult struct {
	File      string
	StartLine int
	Line      int
	Content   string
}

// Grep performs a naive "code search" via git grep
func (r Repo) Grep(search string) ([]GrepResult, error) {
	// Plain-text search only
	re, err := regexp.Compile(regexp.QuoteMeta(search))
	if err != nil {
		return nil, err
	}

	repo, err := r.Git()
	if err != nil {
		return nil, err
	}

	// Loosely modifed from
	// https://github.com/go-git/go-git/blob/fb04aa392c8d4c259cb5b21c1cb4c6f8076e600b/options.go#L736-L740
	// https://github.com/go-git/go-git/blob/fb04aa392c8d4c259cb5b21c1cb4c6f8076e600b/worktree.go#L753-L760
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	return findMatchInFiles(tree.Files(), ref.Hash().String(), &git.GrepOptions{
		Patterns: []*regexp.Regexp{re},
	})
}

// Lines below are copied and modifed from https://github.com/go-git/go-git/blob/fb04aa392c8d4c259cb5b21c1cb4c6f8076e600b/worktree.go#L961-L1045

// findMatchInFiles takes a FileIter, worktree name and GrepOptions, and
// returns a slice of GrepResult containing the result of regex pattern matching
// in content of all the files.
func findMatchInFiles(fileiter *object.FileIter, treeName string, opts *git.GrepOptions) ([]GrepResult, error) {
	var results []GrepResult

	err := fileiter.ForEach(func(file *object.File) error {
		var fileInPathSpec bool

		// When no pathspecs are provided, search all the files.
		if len(opts.PathSpecs) == 0 {
			fileInPathSpec = true
		}

		// Check if the file name matches with the pathspec. Break out of the
		// loop once a match is found.
		for _, pathSpec := range opts.PathSpecs {
			if pathSpec != nil && pathSpec.MatchString(file.Name) {
				fileInPathSpec = true
				break
			}
		}

		// If the file does not match with any of the pathspec, skip it.
		if !fileInPathSpec {
			return nil
		}

		grepResults, err := findMatchInFile(file, treeName, opts)
		if err != nil {
			return err
		}
		results = append(results, grepResults...)

		return nil
	})

	return results, err
}

// findMatchInFile takes a single File, worktree name and GrepOptions,
// and returns a slice of GrepResult containing the result of regex pattern
// matching in the given file.
func findMatchInFile(file *object.File, treeName string, opts *git.GrepOptions) ([]GrepResult, error) {
	var grepResults []GrepResult

	content, err := file.Contents()
	if err != nil {
		return grepResults, err
	}

	// Split the file content and parse line-by-line.
	contentByLine := strings.Split(content, "\n")
	for lineNum, cnt := range contentByLine {
		addToResult := false

		// Match the patterns and content. Break out of the loop once a
		// match is found.
		for _, pattern := range opts.Patterns {
			if pattern != nil && pattern.MatchString(cnt) {
				// Add to result only if invert match is not enabled.
				if !opts.InvertMatch {
					addToResult = true
					break
				}
			} else if opts.InvertMatch {
				// If matching fails, and invert match is enabled, add to
				// results.
				addToResult = true
				break
			}
		}

		if addToResult {
			startLine := lineNum + 1
			ctx := []string{cnt}
			if lineNum != 0 {
				startLine -= 1
				ctx = append([]string{contentByLine[lineNum-1]}, ctx...)
			}
			if lineNum != len(contentByLine)-1 {
				ctx = append(ctx, contentByLine[lineNum+1])
			}
			grepResults = append(grepResults, GrepResult{
				File:      file.Name,
				StartLine: startLine,
				Line:      lineNum + 1,
				Content:   strings.Join(ctx, "\n"),
			})
		}
	}

	return grepResults, nil
}
