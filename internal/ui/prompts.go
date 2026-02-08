package ui

import "github.com/charmbracelet/huh"

type Prompter interface {
	SelectBranch(branches []string) (string, error)
	SelectWorktree(worktrees []string) (string, error)
	SelectEditor(editors []string) (string, error)
	Confirm(title string) (bool, error)
	InputString(title, placeholder string) (string, error)
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
		Options(opts...).
		Height(15).
		Value(&selected)

	err := huh.NewForm(huh.NewGroup(field)).
		WithTheme(WtTheme()).
		WithShowHelp(false).
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
		Options(opts...).
		Height(15).
		Value(&selected)

	err := huh.NewForm(huh.NewGroup(field)).
		WithTheme(WtTheme()).
		WithShowHelp(false).
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
		Options(opts...).
		Height(15).
		Value(&selected)

	err := huh.NewForm(huh.NewGroup(field)).
		WithTheme(WtTheme()).
		WithShowHelp(false).
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
		WithShowHelp(false).
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
		WithShowHelp(false).
		Run()

	return value, err
}
