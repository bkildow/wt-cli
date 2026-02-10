package ui

import (
	"errors"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

type Prompter interface {
	SelectBranch(branches []string) (string, error)
	SelectWorktree(worktrees []string) (string, error)
	SelectEditor(editors []string) (string, error)
	Confirm(title string) (bool, error)
	InputString(title, placeholder string) (string, error)
}

// IsUserAbort returns true if the error is a user cancellation (ESC/ctrl+c).
func IsUserAbort(err error) bool {
	return errors.Is(err, huh.ErrUserAborted)
}

func wtKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	return km
}

func selectHeight(itemCount int) int {
	if itemCount <= 20 {
		return 0
	}
	return 22
}

type InteractivePrompter struct{}

func (p *InteractivePrompter) SelectBranch(branches []string) (string, error) {
	var selected string
	opts := make([]huh.Option[string], len(branches))
	for i, b := range branches {
		opts[i] = huh.NewOption(b, b)
	}

	field := huh.NewSelect[string]().
		Title("Select a branch").
		Options(opts...)

	if h := selectHeight(len(opts)); h > 0 {
		field = field.Height(h)
	}
	field = field.Value(&selected)

	err := huh.NewForm(huh.NewGroup(field)).
		WithTheme(WtTheme()).
		WithKeyMap(wtKeyMap()).
		WithShowHelp(true).
		Run()

	return selected, err
}

func (p *InteractivePrompter) SelectWorktree(worktrees []string) (string, error) {
	var selected string
	opts := make([]huh.Option[string], len(worktrees))
	for i, w := range worktrees {
		opts[i] = huh.NewOption(w, w)
	}

	field := huh.NewSelect[string]().
		Title("Select a worktree").
		Options(opts...)

	if h := selectHeight(len(opts)); h > 0 {
		field = field.Height(h)
	}
	field = field.Value(&selected)

	err := huh.NewForm(huh.NewGroup(field)).
		WithTheme(WtTheme()).
		WithKeyMap(wtKeyMap()).
		WithShowHelp(true).
		Run()

	return selected, err
}

func (p *InteractivePrompter) SelectEditor(editors []string) (string, error) {
	var selected string
	opts := make([]huh.Option[string], len(editors))
	for i, e := range editors {
		opts[i] = huh.NewOption(e, e)
	}

	field := huh.NewSelect[string]().
		Title("Select an editor").
		Options(opts...)

	if h := selectHeight(len(opts)); h > 0 {
		field = field.Height(h)
	}
	field = field.Value(&selected)

	err := huh.NewForm(huh.NewGroup(field)).
		WithTheme(WtTheme()).
		WithKeyMap(wtKeyMap()).
		WithShowHelp(true).
		Run()

	return selected, err
}

func (p *InteractivePrompter) Confirm(title string) (bool, error) {
	var confirmed bool
	field := huh.NewConfirm().
		Title(title).
		Value(&confirmed)

	err := huh.NewForm(huh.NewGroup(field)).
		WithTheme(WtTheme()).
		WithKeyMap(wtKeyMap()).
		WithShowHelp(true).
		Run()

	return confirmed, err
}

func (p *InteractivePrompter) InputString(title, placeholder string) (string, error) {
	var value string
	field := huh.NewInput().
		Title(title).
		Placeholder(placeholder).
		Value(&value)

	err := huh.NewForm(huh.NewGroup(field)).
		WithTheme(WtTheme()).
		WithKeyMap(wtKeyMap()).
		WithShowHelp(true).
		Run()

	return value, err
}
