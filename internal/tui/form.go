package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.jolheiser.com/ugit/internal/git"
)

type repoForm struct {
	inputs       []textinput.Model
	isPrivate    bool
	focusIndex   int
	width        int
	height       int
	done         bool
	save         bool
	selectedRepo *git.Repo
}

// newRepoForm creates a new repository editing form
func newRepoForm() repoForm {
	var inputs []textinput.Model

	nameInput := textinput.New()
	nameInput.Placeholder = "Repository name"
	nameInput.Focus()
	nameInput.Width = 50
	inputs = append(inputs, nameInput)

	descInput := textinput.New()
	descInput.Placeholder = "Repository description"
	descInput.Width = 50
	inputs = append(inputs, descInput)

	tagsInput := textinput.New()
	tagsInput.Placeholder = "Tags (comma separated)"
	tagsInput.Width = 50
	inputs = append(inputs, tagsInput)

	return repoForm{
		inputs:     inputs,
		focusIndex: 0,
	}
}

// setValues sets the form values from the selected repo
func (f *repoForm) setValues(repo *git.Repo) {
	f.inputs[0].SetValue(repo.Name())
	f.inputs[1].SetValue(repo.Meta.Description)
	f.inputs[2].SetValue(strings.Join(repo.Meta.Tags.Slice(), ", "))
	f.isPrivate = repo.Meta.Private

	f.inputs[0].Focus()
	f.focusIndex = 0
}

// setSize sets the form dimensions
func (f *repoForm) setSize(width, height int) {
	f.width = width
	f.height = height

	for i := range f.inputs {
		f.inputs[i].Width = width - 10
	}
}

// isPrivateToggleFocused returns true if the private toggle is focused
func (f *repoForm) isPrivateToggleFocused() bool {
	return f.focusIndex == len(f.inputs)
}

// isSaveButtonFocused returns true if the save button is focused
func (f *repoForm) isSaveButtonFocused() bool {
	return f.focusIndex == len(f.inputs)+1
}

// isCancelButtonFocused returns true if the cancel button is focused
func (f *repoForm) isCancelButtonFocused() bool {
	return f.focusIndex == len(f.inputs)+2
}

// Update handles form updates
func (f repoForm) Update(msg tea.Msg) (repoForm, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			if msg.String() == "up" || msg.String() == "shift+tab" {
				f.focusIndex--
				if f.focusIndex < 0 {
					f.focusIndex = len(f.inputs) + 3 - 1
				}
			} else {
				f.focusIndex++
				if f.focusIndex >= len(f.inputs)+3 {
					f.focusIndex = 0
				}
			}

			for i := range f.inputs {
				if i == f.focusIndex {
					cmds = append(cmds, f.inputs[i].Focus())
				} else {
					f.inputs[i].Blur()
				}
			}

		case "enter":
			if f.isSaveButtonFocused() {
				f.done = true
				f.save = true
				return f, nil
			}

			if f.isCancelButtonFocused() {
				f.done = true
				f.save = false
				return f, nil
			}

		case "esc":
			f.done = true
			f.save = false
			return f, nil

		case " ":
			if f.isPrivateToggleFocused() {
				f.isPrivate = !f.isPrivate
			}

			if f.isSaveButtonFocused() {
				f.done = true
				f.save = true
				return f, nil
			}

			if f.isCancelButtonFocused() {
				f.done = true
				f.save = false
				return f, nil
			}
		}
	}

	for i := range f.inputs {
		if i == f.focusIndex {
			var cmd tea.Cmd
			f.inputs[i], cmd = f.inputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return f, tea.Batch(cmds...)
}

// View renders the form
func (f repoForm) View() string {
	var b strings.Builder

	formStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Edit Repository"))
	b.WriteString("\n\n")

	b.WriteString("Repository Name:\n")
	b.WriteString(f.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString("Description:\n")
	b.WriteString(f.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString("Tags (comma separated):\n")
	b.WriteString(f.inputs[2].View())
	b.WriteString("\n\n")

	toggleStyle := lipgloss.NewStyle()
	if f.isPrivateToggleFocused() {
		toggleStyle = toggleStyle.Foreground(lipgloss.Color("170")).Bold(true)
	}

	visibility := "Public 🔓"
	if f.isPrivate {
		visibility = "Private 🔒"
	}

	b.WriteString(toggleStyle.Render(fmt.Sprintf("[%s] %s", visibility, "Toggle with Space")))
	b.WriteString("\n\n")

	buttonStyle := lipgloss.NewStyle().
		Padding(0, 3).
		MarginRight(1)

	focusedButtonStyle := buttonStyle.Copy().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("170")).
		Bold(true)

	saveButton := buttonStyle.Render("[ Save ]")
	cancelButton := buttonStyle.Render("[ Cancel ]")

	if f.isSaveButtonFocused() {
		saveButton = focusedButtonStyle.Render("[ Save ]")
	}

	if f.isCancelButtonFocused() {
		cancelButton = focusedButtonStyle.Render("[ Cancel ]")
	}

	b.WriteString(saveButton + cancelButton)
	b.WriteString("\n\n")

	b.WriteString("\nTab: Next • Shift+Tab: Previous • Enter: Select • Esc: Cancel")

	return formStyle.Width(f.width - 4).Render(b.String())
}
