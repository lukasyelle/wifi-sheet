package wifiqr

import (
	"strings"
	"testing"
)

func TestPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ssid     string
		password string
		hidden   bool
		want     string
	}{
		{
			name:     "wpa visible",
			ssid:     "Home",
			password: "secret",
			want:     "WIFI:T:WPA;S:Home;P:secret;H:false;;",
		},
		{
			name:     "open network omits password",
			ssid:     "Cafe",
			password: "",
			want:     "WIFI:T:nopass;S:Cafe;H:false;;",
		},
		{
			name:     "hidden network",
			ssid:     "Home",
			password: "secret",
			hidden:   true,
			want:     "WIFI:T:WPA;S:Home;P:secret;H:true;;",
		},
		{
			name:     "escapes special characters",
			ssid:     `Net;work:1`,
			password: `p\ass,word"`,
			want:     `WIFI:T:WPA;S:Net\;work\:1;P:p\\ass\,word\";H:false;;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Payload(tt.ssid, tt.password, tt.hidden)
			if got != tt.want {
				t.Fatalf("Payload() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPNG(t *testing.T) {
	t.Parallel()

	png, err := PNG("Home", "secret", false, 512)
	if err != nil {
		t.Fatalf("PNG() error = %v", err)
	}
	if len(png) < 100 {
		t.Fatalf("PNG() too small: %d bytes", len(png))
	}
	if string(png[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Fatalf("PNG() missing PNG magic")
	}
}

func TestPayloadDoesNotEmbedRawSemicolonInSSID(t *testing.T) {
	t.Parallel()
	got := Payload("a;b", "x", false)
	if strings.Contains(got, ";S:a;b;") {
		t.Fatalf("unescaped semicolon in SSID: %q", got)
	}
}
