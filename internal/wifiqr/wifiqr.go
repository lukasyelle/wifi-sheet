package wifiqr

import (
	"fmt"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

const (
	authWPA    = "WPA"
	authNopass = "nopass"
)

// Payload builds a WIFI: MeCard string that iOS and Android will join on scan.
func Payload(ssid, password string, hidden bool) string {
	auth := authWPA
	if password == "" {
		auth = authNopass
	}

	hiddenValue := "false"
	if hidden {
		hiddenValue = "true"
	}

	var b strings.Builder
	b.WriteString("WIFI:T:")
	b.WriteString(auth)
	b.WriteString(";S:")
	b.WriteString(escape(ssid))
	if password != "" {
		b.WriteString(";P:")
		b.WriteString(escape(password))
	}
	b.WriteString(";H:")
	b.WriteString(hiddenValue)
	b.WriteString(";;")
	return b.String()
}

// PNG returns a high-contrast QR image with a quiet zone, sized for print.
func PNG(ssid, password string, hidden bool, pixels int) ([]byte, error) {
	if pixels < 256 {
		pixels = 1024
	}
	payload := Payload(ssid, password, hidden)
	png, err := qrcode.Encode(payload, qrcode.High, pixels)
	if err != nil {
		return nil, fmt.Errorf("encode wifi qr: %w", err)
	}
	return png, nil
}

func escape(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`;`, `\;`,
		`,`, `\,`,
		`:`, `\:`,
		`"`, `\"`,
	)
	return replacer.Replace(s)
}
