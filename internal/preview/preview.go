package preview

import (
	"fmt"
	"os/exec"
	"strings"
)

// Open launches paths in the macOS default app (Preview for PDFs).
func Open(paths ...string) error {
	if len(paths) == 0 {
		return nil
	}
	cmd := exec.Command("open", paths...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("open preview: %w (%s)", err, msg)
		}
		return fmt.Errorf("open preview: %w", err)
	}
	return nil
}
