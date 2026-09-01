package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/term"

	"github.com/lukasyelle/wifi-sheet/assets"
	"github.com/lukasyelle/wifi-sheet/internal/networks"
	"github.com/lukasyelle/wifi-sheet/internal/prompt"
	"github.com/lukasyelle/wifi-sheet/internal/sheet"
	"github.com/lukasyelle/wifi-sheet/internal/wifiqr"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "wifi-sheet: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, in io.Reader, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("wifi-sheet", flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.Usage = func() {
		fmt.Fprintf(errOut, "Usage: wifi-sheet [-o dir]\n\n")
		fmt.Fprintf(errOut, "Generate US Letter PDFs with a scannable WiFi QR code.\n")
		fs.PrintDefaults()
	}
	outDir := fs.String("o", ".", "directory to write PDF files")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	ssids, err := networks.List()
	if err != nil {
		return err
	}
	if len(ssids) == 0 {
		return fmt.Errorf("no saved WiFi networks found")
	}

	fmt.Fprintln(out, "Saved WiFi networks")
	fmt.Fprintln(out)
	for i, ssid := range ssids {
		fmt.Fprintf(out, "  %2d) %s\n", i+1, ssid)
	}
	fmt.Fprintln(out)
	fmt.Fprint(out, "Select one or more (1,3 or all): ")

	line, err := prompt.ReadLine(in)
	if err != nil {
		return err
	}
	indexes, err := prompt.ParseSelection(line, len(ssids))
	if err != nil {
		return err
	}

	if len(indexes) > 1 {
		fmt.Fprintf(errOut, "Generating %d sheets. Allow Keychain access if macOS asks.\n", len(indexes))
	}

	written := 0
	for _, idx := range indexes {
		ssid := ssids[idx]
		password, err := passwordFor(ssid, in, out, errOut)
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

		path, err := sheet.UniquePath(*outDir, sheet.Filename(ssid))
		if err != nil {
			return fmt.Errorf("%s: %w", ssid, err)
		}
		if err := sheet.WriteFile(path, ssid, qrPNG, assets.Background); err != nil {
			return fmt.Errorf("%s: %w", ssid, err)
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			abs = path
		}
		fmt.Fprintf(out, "Wrote %s\n", abs)
		written++
	}

	if written > 1 {
		fmt.Fprintf(out, "Wrote %d sheets.\n", written)
	}
	return nil
}

func passwordFor(ssid string, in io.Reader, out, errOut io.Writer) (string, error) {
	password, err := networks.KeychainPassword(ssid)
	if err == nil {
		return password, nil
	}
	if errors.Is(err, networks.ErrNotFound) {
		fmt.Fprintf(errOut, "No saved password for %s. Type it to continue.\n", ssid)
	} else {
		fmt.Fprintf(errOut, "Could not read the Keychain password for %s.\nType it to continue.\n", ssid)
	}
	return readPassword(ssid, in, out)
}

func readPassword(ssid string, in io.Reader, out io.Writer) (string, error) {
	fmt.Fprintf(out, "Password for %s (input hidden): ", ssid)

	file, ok := in.(*os.File)
	if ok && term.IsTerminal(int(file.Fd())) {
		secret, err := term.ReadPassword(int(file.Fd()))
		fmt.Fprintln(out)
		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				return "", fmt.Errorf("read password: interrupted")
			}
			return "", fmt.Errorf("read password: %w", err)
		}
		return string(secret), nil
	}

	line, err := prompt.ReadLine(in)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return line, nil
}
