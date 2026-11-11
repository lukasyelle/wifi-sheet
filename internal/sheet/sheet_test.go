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
		name  string
		ssids []string
		want  string
	}{
		{name: "single", ssids: []string{"Home"}, want: "wifi-Home.pdf"},
		{name: "spaces", ssids: []string{"NSA Listening Post"}, want: "wifi-NSA-Listening-Post.pdf"},
		{name: "trimmed", ssids: []string{"  Guest  WiFi  "}, want: "wifi-Guest-WiFi.pdf"},
		{name: "symbols", ssids: []string{"Net/Work:1"}, want: "wifi-NetWork1.pdf"},
		{name: "empty-ish", ssids: []string{"!!!"}, want: "wifi-network.pdf"},
		{name: "two", ssids: []string{"Home", "Guest"}, want: "wifi-Home-Guest.pdf"},
		{name: "none", ssids: nil, want: "wifi-networks.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Filename(tt.ssids...); got != tt.want {
				t.Fatalf("Filename(%v) = %q, want %q", tt.ssids, got, tt.want)
			}
		})
	}
}

func TestGridDims(t *testing.T) {
	t.Parallel()

	tests := []struct {
		n        int
		wantCols int
		wantRows int
	}{
		{1, 1, 1},
		{2, 2, 1},
		{3, 2, 2},
		{4, 2, 2},
		{5, 2, 3},
		{6, 2, 3},
		{7, 3, 3},
		{9, 3, 3},
		{10, 3, 4},
	}
	for _, tt := range tests {
		cols, rows := gridDims(tt.n)
		if cols != tt.wantCols || rows != tt.wantRows {
			t.Fatalf("gridDims(%d) = %d,%d want %d,%d", tt.n, cols, rows, tt.wantCols, tt.wantRows)
		}
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
	if err := WriteFile(path, []Network{{SSID: "Home", QRPNG: qr}}, assets.Background); err != nil {
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

func TestWriteFileMultiple(t *testing.T) {
	t.Parallel()

	homeQR, err := wifiqr.PNG("Home", "home-secret", false, 512)
	if err != nil {
		t.Fatalf("wifiqr.PNG(Home) error = %v", err)
	}
	guestQR, err := wifiqr.PNG("Guest", "guest-secret", false, 512)
	if err != nil {
		t.Fatalf("wifiqr.PNG(Guest) error = %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, Filename("Home", "Guest"))
	err = WriteFile(path, []Network{
		{SSID: "Home", QRPNG: homeQR},
		{SSID: "Guest", QRPNG: guestQR},
	}, assets.Background)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Fatalf("WriteFile() did not produce a PDF")
	}
	if bytes.Contains(data, []byte("/Count 2")) {
		t.Fatalf("WriteFile() looks like more than one page")
	}
	if bytes.Contains(data, []byte("home-secret")) || bytes.Contains(data, []byte("guest-secret")) {
		t.Fatal("password leaked into the PDF as text")
	}
}

func TestWriteFileRequiresNetwork(t *testing.T) {
	t.Parallel()
	err := WriteFile(filepath.Join(t.TempDir(), "empty.pdf"), nil, assets.Background)
	if err == nil {
		t.Fatal("WriteFile() expected error for empty networks")
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

	homeQR, err := wifiqr.PNG("Home", "example-password", false, 1200)
	if err != nil {
		t.Fatalf("wifiqr.PNG() error = %v", err)
	}
	guestQR, err := wifiqr.PNG("Guest", "guest-password", false, 1200)
	if err != nil {
		t.Fatalf("wifiqr.PNG() error = %v", err)
	}

	single := filepath.Join(dir, "wifi-Home.pdf")
	if err := WriteFile(single, []Network{{SSID: "Home", QRPNG: homeQR}}, assets.Background); err != nil {
		t.Fatalf("WriteFile(single) error = %v", err)
	}

	multi := filepath.Join(dir, "wifi-Home-Guest.pdf")
	err = WriteFile(multi, []Network{
		{SSID: "Home", QRPNG: homeQR},
		{SSID: "Guest", QRPNG: guestQR},
	}, assets.Background)
	if err != nil {
		t.Fatalf("WriteFile(multi) error = %v", err)
	}
}
