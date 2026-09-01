package networks

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrNotFound is returned when Keychain has no AirPort password for the SSID.
var ErrNotFound = errors.New("wifi password not in keychain")

// List returns preferred WiFi SSIDs on the Mac, with the current network first
// when it can be determined.
func List() ([]string, error) {
	device, err := wifiDevice()
	if err != nil {
		return nil, err
	}

	preferred, err := preferredNetworks(device)
	if err != nil {
		return nil, err
	}

	current, err := currentSSID(device)
	if err != nil {
		return nil, err
	}

	return mergeCurrent(preferred, current), nil
}

// KeychainPassword reads the AirPort password for ssid from the login keychain.
// macOS will prompt the user to allow access.
func KeychainPassword(ssid string) (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-D", "AirPort network password",
		"-w",
		"-a", ssid,
	)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 44 {
			return "", ErrNotFound
		}
		stderr := ""
		if errors.As(err, &exitErr) {
			stderr = strings.TrimSpace(string(exitErr.Stderr))
		}
		if stderr != "" {
			return "", fmt.Errorf("keychain lookup for %q: %w (%s)", ssid, err, stderr)
		}
		return "", fmt.Errorf("keychain lookup for %q: %w", ssid, err)
	}

	password := strings.TrimRight(string(out), "\n")
	if password == "" {
		return "", ErrNotFound
	}
	return password, nil
}

func wifiDevice() (string, error) {
	out, err := exec.Command("networksetup", "-listallhardwareports").Output()
	if err != nil {
		return "", fmt.Errorf("list hardware ports: %w", err)
	}
	device, err := parseWifiDevice(string(out))
	if err != nil {
		return "", err
	}
	return device, nil
}

func parseWifiDevice(output string) (string, error) {
	blocks := strings.Split(output, "\n\n")
	for _, block := range blocks {
		if !strings.Contains(block, "Hardware Port: Wi-Fi") &&
			!strings.Contains(block, "Hardware Port: AirPort") {
			continue
		}
		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "Device:") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return "", fmt.Errorf("wifi device: malformed Device line %q", line)
			}
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("wifi device: no Wi-Fi hardware port found")
}

func preferredNetworks(device string) ([]string, error) {
	out, err := exec.Command("networksetup", "-listpreferredwirelessnetworks", device).Output()
	if err != nil {
		return nil, fmt.Errorf("list preferred networks: %w", err)
	}
	return parsePreferredNetworks(string(out)), nil
}

func parsePreferredNetworks(output string) []string {
	var ssids []string
	for i, line := range strings.Split(output, "\n") {
		if i == 0 {
			continue
		}
		ssid := strings.TrimSpace(strings.TrimPrefix(line, "\t"))
		if ssid == "" {
			continue
		}
		ssids = append(ssids, ssid)
	}
	return ssids
}

func currentSSID(device string) (string, error) {
	out, err := exec.Command("ipconfig", "getsummary", device).Output()
	if err != nil {
		return "", nil
	}
	return parseCurrentSSID(string(out)), nil
}

func parseCurrentSSID(output string) string {
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "SSID :") && !strings.HasPrefix(trimmed, "SSID:") {
			continue
		}
		_, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			return ""
		}
		ssid := strings.TrimSpace(value)
		if ssid == "" || ssid == "<redacted>" {
			return ""
		}
		return ssid
	}
	return ""
}

func mergeCurrent(preferred []string, current string) []string {
	if current == "" {
		return dedupe(preferred)
	}

	out := []string{current}
	for _, ssid := range preferred {
		if ssid == current {
			continue
		}
		out = append(out, ssid)
	}
	return dedupe(out)
}

func dedupe(ssids []string) []string {
	seen := make(map[string]struct{}, len(ssids))
	out := make([]string, 0, len(ssids))
	for _, ssid := range ssids {
		if _, ok := seen[ssid]; ok {
			continue
		}
		seen[ssid] = struct{}{}
		out = append(out, ssid)
	}
	return out
}
