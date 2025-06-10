package tui

import (
	"strings"

	"go.jolheiser.com/ugit/internal/git"
)

// repoItem represents a repository item in the list
type repoItem struct {
	repo *git.Repo
}

// Title returns the title for the list item
func (r repoItem) Title() string {
	return r.repo.Name()
}

// Description returns the description for the list item
func (r repoItem) Description() string {
	var builder strings.Builder

	if r.repo.Meta.Private {
		builder.WriteString("🔒")
	} else {
		builder.WriteString("🔓")
	}

	builder.WriteString(" • ")

	if r.repo.Meta.Description != "" {
		builder.WriteString(r.repo.Meta.Description)
	} else {
		builder.WriteString("No description")
	}

	builder.WriteString(" • ")

	builder.WriteString("[")
	if len(r.repo.Meta.Tags) > 0 {
		builder.WriteString(strings.Join(r.repo.Meta.Tags.Slice(), ", "))
	}
	builder.WriteString("]")

	builder.WriteString(" • ")

	lastCommit, err := r.repo.LastCommit()
	if err == nil {
		builder.WriteString(lastCommit.Short())
	} else {
		builder.WriteString("deadbeef")
	}

	return builder.String()
}

// FilterValue returns the value to use for filtering
func (r repoItem) FilterValue() string {
	var builder strings.Builder
	builder.WriteString(r.repo.Name())
	builder.WriteString(" ")
	builder.WriteString(r.repo.Meta.Description)

	if len(r.repo.Meta.Tags) > 0 {
		for _, tag := range r.repo.Meta.Tags.Slice() {
			builder.WriteString(" ")
			builder.WriteString(tag)
		}
	}

	return strings.ToLower(builder.String())
}
