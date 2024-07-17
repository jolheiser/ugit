package ssh

import (
	"strings"

	"github.com/charmbracelet/huh"
	"go.jolheiser.com/ugit/internal/git"
)

type stringSliceAccessor []string

func (s stringSliceAccessor) Get() string {
	return strings.Join(s, "\n")
}

func (s stringSliceAccessor) Set(value string) {
	s = strings.Split(value, "\n")
}

func newForm(meta git.RepoMeta) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Description").
				Value(&meta.Description),
			huh.NewConfirm().
				Title("Visibility").
				Affirmative("Public").
				Negative("Private").
				Value(&meta.Private),
			huh.NewText().
				Title("Tags").
				Description("One per line").
				Accessor(stringSliceAccessor(meta.Tags)),
		),
	)
}
