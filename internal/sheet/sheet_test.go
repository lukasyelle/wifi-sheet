package sheet

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lukasyelle/wifi-sheet/assets"
	"github.com/lukasyelle/wifi-sheet/internal/wifiqr"
)

func TestFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ssid string
		want string
	}{
		{ssid: "Home", want: "wifi-Home.pdf"},
		{ssid: "NSA Listening Post", want: "wifi-NSA-Listening-Post.pdf"},
		{ssid: "  Guest  WiFi  ", want: "wifi-Guest-WiFi.pdf"},
		{ssid: "Net/Work:1", want: "wifi-NetWork1.pdf"},
		{ssid: "!!!", want: "wifi-network.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.ssid, func(t *testing.T) {
			t.Parallel()
			if got := Filename(tt.ssid); got != tt.want {
				t.Fatalf("Filename(%q) = %q, want %q", tt.ssid, got, tt.want)
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	t.Parallel()

	qr, err := wifiqr.PNG("Home", "test-password", false, 512)
	if err != nil {
		t.Fatalf("wifiqr.PNG() error = %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, Filename("Home"))
	if err := WriteFile(path, "Home", qr, assets.Background); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Fatalf("WriteFile() did not produce a PDF")
	}
	if !bytes.Contains(data, []byte("/Type /Page")) && !bytes.Contains(data, []byte("/Type/Page")) {
		t.Fatalf("WriteFile() PDF missing page object")
	}
	if bytes.Contains(data, []byte("/Count 2")) {
		t.Fatalf("WriteFile() looks like more than one page")
	}
	if bytes.Contains(data, []byte("test-password")) {
		t.Fatal("password leaked into the PDF as text")
	}
}

func TestUniquePath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	first, err := UniquePath(dir, "wifi-Home.pdf")
	if err != nil {
		t.Fatalf("UniquePath() error = %v", err)
	}
	if err := os.WriteFile(first, []byte("a"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	second, err := UniquePath(dir, "wifi-Home.pdf")
	if err != nil {
		t.Fatalf("UniquePath() error = %v", err)
	}
	if filepath.Base(second) != "wifi-Home-2.pdf" {
		t.Fatalf("UniquePath() = %q, want wifi-Home-2.pdf", filepath.Base(second))
	}
}

func TestWriteFilePreview(t *testing.T) {
	dir := os.Getenv("WIFI_SHEET_PREVIEW")
	if dir == "" {
		t.Skip("set WIFI_SHEET_PREVIEW to write a sample PDF")
	}

	qr, err := wifiqr.PNG("Home", "example-password", false, 1200)
	if err != nil {
		t.Fatalf("wifiqr.PNG() error = %v", err)
	}
	path := filepath.Join(dir, "wifi-Home.pdf")
	if err := WriteFile(path, "Home", qr, assets.Background); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
