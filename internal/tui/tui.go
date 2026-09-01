package tui

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
)

// ErrAborted is returned when the user quits the form with ctrl+c.
var ErrAborted = huh.ErrUserAborted

// SelectNetworks shows a filterable multi-select list of SSIDs.
func SelectNetworks(ssids []string) ([]string, error) {
	if len(ssids) == 0 {
		return nil, fmt.Errorf("no networks to select")
	}

	var selected []string
	field := huh.NewMultiSelect[string]().
		Title("Saved WiFi networks").
		Description("Type / to filter. Space toggles. Enter generates the sheet.").
		Options(huh.NewOptions(ssids...)...).
		Value(&selected).
		Filterable(true).
		Height(16).
		Validate(func(values []string) error {
			if len(values) == 0 {
				return fmt.Errorf("select at least one network")
			}
			return nil
		})

	form := huh.NewForm(huh.NewGroup(field)).WithTheme(huh.ThemeCharm())
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrAborted
		}
		return nil, fmt.Errorf("select networks: %w", err)
	}
	return selected, nil
}

// Password asks for a hidden password when Keychain does not have one.
func Password(ssid string) (string, error) {
	var value string
	field := huh.NewInput().
		Title(fmt.Sprintf("Password for %s", ssid)).
		Description("Not saved in Keychain. Input is hidden. Leave blank for an open network.").
		Password(true).
		Value(&value)

	form := huh.NewForm(huh.NewGroup(field)).WithTheme(huh.ThemeCharm())
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrAborted
		}
		return "", fmt.Errorf("read password: %w", err)
	}
	return value, nil
}
