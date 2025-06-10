package tui

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"go.jolheiser.com/ugit/internal/git"
)

// Model is the main TUI model
type Model struct {
	repoList   list.Model
	repos      []*git.Repo
	repoDir    string
	width      int
	height     int
	help       help.Model
	keys       keyMap
	activeView View
	repoForm   repoForm
	session    ssh.Session
}

// View represents the current active view in the TUI
type View int

const (
	ViewList View = iota
	ViewForm
	ViewConfirmDelete
)

// New creates a new TUI model
func New(s ssh.Session, repoDir string) (*Model, error) {
	repos, err := loadRepos(repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load repos: %w", err)
	}

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

	repoList.FilterInput.Placeholder = "Type to filter repositories..."
	repoList.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	repoList.FilterInput.TextStyle = lipgloss.NewStyle()

	help := help.New()

	repoForm := newRepoForm()

	return &Model{
		repoList:   repoList,
		repos:      repos,
		repoDir:    repoDir,
		help:       help,
		keys:       keys,
		activeView: ViewList,
		repoForm:   repoForm,
		session:    s,
	}, nil
}

// loadRepos loads all git repositories from the given directory
func loadRepos(repoDir string) ([]*git.Repo, error) {
	entries, err := git.ListRepos(repoDir)
	if err != nil {
		return nil, err
	}

	repos := make([]*git.Repo, 0, len(entries))
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".git") {
			continue
		}
		repo, err := git.NewRepo(repoDir, entry.Name())
		if err != nil {
			slog.Error("error loading repo", "name", entry.Name(), "error", err)
			continue
		}
		repos = append(repos, repo)
	}

	return repos, nil
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles all the messages and updates the model accordingly
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}

		switch m.activeView {
		case ViewList:
			var cmd tea.Cmd
			m.repoList, cmd = m.repoList.Update(msg)
			cmds = append(cmds, cmd)

			if m.repoList.FilterState() == list.Filtering {
				break
			}

			switch {
			case key.Matches(msg, m.keys.Edit):
				if len(m.repos) == 0 {
					m.repoList.NewStatusMessage("No repositories to edit")
					break
				}

				selectedItem := m.repoList.SelectedItem().(repoItem)
				m.repoForm.selectedRepo = selectedItem.repo

				m.repoForm.setValues(selectedItem.repo)
				m.activeView = ViewForm
				return m, textinput.Blink

			case key.Matches(msg, m.keys.Delete):
				if len(m.repos) == 0 {
					m.repoList.NewStatusMessage("No repositories to delete")
					break
				}

				m.activeView = ViewConfirmDelete
			}

		case ViewForm:
			var cmd tea.Cmd
			m.repoForm, cmd = m.repoForm.Update(msg)
			cmds = append(cmds, cmd)

			if m.repoForm.done {
				if m.repoForm.save {
					selectedRepo := m.repoForm.selectedRepo
					repoDir := filepath.Dir(selectedRepo.Path())
					oldName := selectedRepo.Name()
					newName := m.repoForm.inputs[0].Value()

					var renamed bool
					if oldName != newName {
						if err := git.RenameRepo(repoDir, oldName, newName); err != nil {
							m.repoList.NewStatusMessage(fmt.Sprintf("Error renaming repo: %s", err))
						} else {
							m.repoList.NewStatusMessage(fmt.Sprintf("Repository renamed from %s to %s", oldName, newName))
							renamed = true
						}
					}

					if renamed {
						if newRepo, err := git.NewRepo(repoDir, newName+".git"); err == nil {
							selectedRepo = newRepo
						} else {
							m.repoList.NewStatusMessage(fmt.Sprintf("Error loading renamed repo: %s", err))
						}
					}

					selectedRepo.Meta.Description = m.repoForm.inputs[1].Value()
					selectedRepo.Meta.Private = m.repoForm.isPrivate

					tags := make(git.TagSet)
					for _, tag := range strings.Split(m.repoForm.inputs[2].Value(), ",") {
						tag = strings.TrimSpace(tag)
						if tag != "" {
							tags.Add(tag)
						}
					}
					selectedRepo.Meta.Tags = tags

					if err := selectedRepo.SaveMeta(); err != nil {
						m.repoList.NewStatusMessage(fmt.Sprintf("Error saving repo metadata: %s", err))
					} else if !renamed {
						m.repoList.NewStatusMessage("Repository updated successfully")
					}
				}

				m.repoForm.done = false
				m.repoForm.save = false
				m.activeView = ViewList

				if repos, err := loadRepos(m.repoDir); err == nil {
					m.repos = repos
					items := make([]list.Item, len(repos))
					for i, repo := range repos {
						items[i] = repoItem{repo: repo}
					}
					m.repoList.SetItems(items)
				}
			}

		case ViewConfirmDelete:
			switch {
			case key.Matches(msg, m.keys.Confirm):
				selectedItem := m.repoList.SelectedItem().(repoItem)
				repo := selectedItem.repo

				if err := git.DeleteRepo(repo.Path()); err != nil {
					m.repoList.NewStatusMessage(fmt.Sprintf("Error deleting repo: %s", err))
				} else {
					m.repoList.NewStatusMessage(fmt.Sprintf("Repository %s deleted", repo.Name()))

					if repos, err := loadRepos(m.repoDir); err == nil {
						m.repos = repos
						items := make([]list.Item, len(repos))
						for i, repo := range repos {
							items[i] = repoItem{repo: repo}
						}
						m.repoList.SetItems(items)
					}
				}
				m.activeView = ViewList

			case key.Matches(msg, m.keys.Cancel):
				m.activeView = ViewList
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		footerHeight := 2

		m.repoList.SetSize(msg.Width, msg.Height-headerHeight-footerHeight)
		m.repoForm.setSize(msg.Width, msg.Height)

		m.help.Width = msg.Width
	}

	return m, tea.Batch(cmds...)
}

// View renders the current UI
func (m Model) View() string {
	switch m.activeView {
	case ViewList:
		return fmt.Sprintf("%s\n%s", m.repoList.View(), m.help.View(m.keys))

	case ViewForm:
		return m.repoForm.View()

	case ViewConfirmDelete:
		selectedItem := m.repoList.SelectedItem().(repoItem)
		repo := selectedItem.repo

		confirmStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Padding(1, 2).
			Width(m.width - 4).
			Align(lipgloss.Center)

		confirmText := fmt.Sprintf(
			"Are you sure you want to delete repository '%s'?\n\nThis action cannot be undone!\n\nPress y to confirm or n to cancel.",
			repo.Name(),
		)

		return confirmStyle.Render(confirmText)
	}

	return ""
}

// Start runs the TUI
func Start(s ssh.Session, repoDir string) error {
	model, err := New(s, repoDir)
	if err != nil {
		return err
	}

	// Get terminal dimensions from SSH session if available
	pty, _, isPty := s.Pty()
	if isPty && pty.Window.Width > 0 && pty.Window.Height > 0 {
		// Initialize with correct size
		model.width = pty.Window.Width
		model.height = pty.Window.Height
		
		headerHeight := 3
		footerHeight := 2
		model.repoList.SetSize(pty.Window.Width, pty.Window.Height-headerHeight-footerHeight)
		model.repoForm.setSize(pty.Window.Width, pty.Window.Height)
		model.help.Width = pty.Window.Width
	}

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithInput(s), tea.WithOutput(s))

	_, err = p.Run()
	return err
}
