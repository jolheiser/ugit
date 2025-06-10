package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Run runs the TUI standalone, useful for development or local usage
func Run(repoDir string) error {
	model := Model{
		repoDir:    repoDir,
		help:       help.New(),
		keys:       keys,
		activeView: ViewList,
		repoForm:   newRepoForm(),
	}

	repos, err := loadRepos(repoDir)
	if err != nil {
		return fmt.Errorf("failed to load repos: %w", err)
	}
	model.repos = repos

	items := make([]list.Item, len(repos))
	for i, repo := range repos {
		items[i] = repoItem{repo: repo}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("170"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("244"))

	repoList := list.New(items, delegate, 0, 0)
	repoList.Title = "Git Repositories"
	repoList.SetShowStatusBar(true)
	repoList.SetFilteringEnabled(true)
	repoList.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).Padding(0, 0, 0, 2)
	repoList.StatusMessageLifetime = 3

	model.repoList = repoList

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	_, err = p.Run()
	return err
}
