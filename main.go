package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lukasyelle/wifi-sheet/assets"
	"github.com/lukasyelle/wifi-sheet/internal/networks"
	"github.com/lukasyelle/wifi-sheet/internal/preview"
	"github.com/lukasyelle/wifi-sheet/internal/sheet"
	"github.com/lukasyelle/wifi-sheet/internal/tui"
	"github.com/lukasyelle/wifi-sheet/internal/wifiqr"
	"golang.org/x/term"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, tui.ErrAborted) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "wifi-sheet: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, in io.Reader, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("wifi-sheet", flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.Usage = func() {
		fmt.Fprintf(errOut, "Usage: wifi-sheet [-o dir]\n\n")
		fmt.Fprintf(errOut, "Generate a US Letter PDF with scannable WiFi QR codes.\n")
		fs.PrintDefaults()
	}
	outDir := fs.String("o", ".", "directory to write the PDF file")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if file, ok := in.(*os.File); !ok || !term.IsTerminal(int(file.Fd())) {
		return fmt.Errorf("needs an interactive terminal")
	}

	ssids, err := networks.List()
	if err != nil {
		return err
	}
	if len(ssids) == 0 {
		return fmt.Errorf("no saved WiFi networks found")
	}

	selected, err := tui.SelectNetworks(ssids)
	if err != nil {
		return err
	}

	if len(selected) > 1 {
		fmt.Fprintf(errOut, "Building one sheet with %d networks. Allow Keychain access if macOS asks.\n", len(selected))
	}

	entries := make([]sheet.Network, 0, len(selected))
	for _, ssid := range selected {
		password, err := passwordFor(ssid, errOut)
		if err != nil {
			return err
		}
		if password == "" {
			fmt.Fprintf(errOut, "No password for %s. Encoding as an open network.\n", ssid)
		}

		qrPNG, err := wifiqr.PNG(ssid, password, false, 1200)
		if err != nil {
			return fmt.Errorf("%s: %w", ssid, err)
		}
		entries = append(entries, sheet.Network{SSID: ssid, QRPNG: qrPNG})
	}

	path, err := sheet.UniquePath(*outDir, sheet.Filename(selected...))
	if err != nil {
		return err
	}
	if err := sheet.WriteFile(path, entries, assets.Background); err != nil {
		return err
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	fmt.Fprintf(out, "Wrote %s\n", abs)

	if err := preview.Open(abs); err != nil {
		return err
	}
	return nil
}

func passwordFor(ssid string, errOut io.Writer) (string, error) {
	password, err := networks.KeychainPassword(ssid)
	if err == nil {
		return password, nil
	}
	if errors.Is(err, networks.ErrNotFound) {
		fmt.Fprintf(errOut, "No saved password for %s.\n", ssid)
	} else {
		fmt.Fprintf(errOut, "Could not read the Keychain password for %s.\n", ssid)
	}
	return tui.Password(ssid)
}
