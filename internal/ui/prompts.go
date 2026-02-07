package ui

import "github.com/charmbracelet/huh"

type Prompter interface {
	SelectBranch(branches []string) (string, error)
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

	err := huh.NewSelect[string]().
		Title("Select a branch").
		Options(opts...).
		Height(15).
		Value(&selected).
		Run()

	return selected, err
}

func (p *InteractivePrompter) Confirm(title string) (bool, error) {
	var confirmed bool
	err := huh.NewConfirm().
		Title(title).
		Value(&confirmed).
		Run()

	return confirmed, err
}

func (p *InteractivePrompter) InputString(title, placeholder string) (string, error) {
	var value string
	err := huh.NewInput().
		Title(title).
		Placeholder(placeholder).
		Value(&value).
		Run()

	return value, err
}
