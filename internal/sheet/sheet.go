package sheet

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/go-pdf/fpdf"
)

const (
	pageW    = 8.5
	pageH    = 11.0
	cardW    = 5.20
	cardH    = 6.20
	qrSize   = 3.55
	corner   = 0.20
	subtitle = "Scan to join this WiFi"
	fontName = "ssid"
	fontReg  = "/System/Library/Fonts/Supplemental/Arial.ttf"
	fontBold = "/System/Library/Fonts/Supplemental/Arial Bold.ttf"
)

// WriteFile renders a single US Letter PDF with a nature background, the SSID,
// and a scannable WiFi QR code. The password is not printed.
func WriteFile(path, ssid string, qrPNG, background []byte) error {
	if ssid == "" {
		return fmt.Errorf("ssid is required")
	}
	if len(qrPNG) == 0 {
		return fmt.Errorf("qr image is required")
	}
	if len(background) == 0 {
		return fmt.Errorf("background image is required")
	}

	pdf := fpdf.New("P", "in", "Letter", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	family, err := registerFonts(pdf)
	if err != nil {
		return err
	}

	bgOpt := fpdf.ImageOptions{ImageType: "JPG", ReadDpi: false}
	pdf.RegisterImageOptionsReader("bg", bgOpt, bytes.NewReader(background))
	if pdf.Error() != nil {
		return fmt.Errorf("background image: %w", pdf.Error())
	}
	pdf.ImageOptions("bg", 0, 0, pageW, pageH, false, bgOpt, 0, "")

	cardX := (pageW - cardW) / 2
	cardY := (pageH-cardH)/2 - 0.08

	pdf.SetAlpha(0.16, "Normal")
	pdf.SetFillColor(18, 22, 28)
	pdf.RoundedRect(cardX+0.07, cardY+0.09, cardW, cardH, corner, "1234", "F")

	pdf.SetAlpha(0.94, "Normal")
	pdf.SetFillColor(255, 255, 255)
	pdf.RoundedRect(cardX, cardY, cardW, cardH, corner, "1234", "F")
	pdf.SetAlpha(1, "Normal")

	titleMaxW := cardW - 0.70
	titleSize := fitText(pdf, family, ssid, titleMaxW, 26, 13)
	pdf.SetFont(family, "B", titleSize)
	pdf.SetTextColor(28, 32, 36)
	pdf.SetXY(cardX, cardY+0.38)
	pdf.CellFormat(cardW, 0.42, ssid, "", 0, "C", false, 0, "")

	pdf.SetFont(family, "", 12)
	pdf.SetTextColor(92, 98, 106)
	pdf.SetXY(cardX, cardY+0.86)
	pdf.CellFormat(cardW, 0.28, subtitle, "", 0, "C", false, 0, "")

	ruleY := cardY + 1.22
	pdf.SetDrawColor(220, 223, 226)
	pdf.SetLineWidth(0.008)
	inset := 1.15
	pdf.Line(cardX+inset, ruleY, cardX+cardW-inset, ruleY)

	qrOpt := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
	pdf.RegisterImageOptionsReader("qr", qrOpt, bytes.NewReader(qrPNG))
	if pdf.Error() != nil {
		return fmt.Errorf("qr image: %w", pdf.Error())
	}
	qrX := cardX + (cardW-qrSize)/2
	qrY := cardY + 1.42
	pdf.ImageOptions("qr", qrX, qrY, qrSize, qrSize, false, qrOpt, 0, "")

	if pdf.PageCount() != 1 {
		return fmt.Errorf("expected 1 page, got %d", pdf.PageCount())
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
	}
	if err := pdf.OutputFileAndClose(path); err != nil {
		return fmt.Errorf("write pdf: %w", err)
	}
	return nil
}

// Filename returns a filesystem-safe PDF name for the SSID.
func Filename(ssid string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range ssid {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || unicode.IsSpace(r):
			if lastDash || b.Len() == 0 {
				continue
			}
			b.WriteByte('-')
			lastDash = true
		}
	}
	name := strings.Trim(b.String(), "-")
	if name == "" {
		name = "network"
	}
	return "wifi-" + name + ".pdf"
}

// UniquePath returns path dir/filename, adding -2, -3, ... if the file exists.
func UniquePath(dir, filename string) (string, error) {
	path := filepath.Join(dir, filename)
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return path, nil
	}
	if err != nil {
		return "", fmt.Errorf("stat %s: %w", path, err)
	}

	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	for i := 2; i < 1000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, i, ext))
		_, err := os.Stat(candidate)
		if errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
		if err != nil {
			return "", fmt.Errorf("stat %s: %w", candidate, err)
		}
	}
	return "", fmt.Errorf("too many existing files for %s", filename)
}

func registerFonts(pdf *fpdf.Fpdf) (string, error) {
	regular, err := os.ReadFile(fontReg)
	if err != nil {
		return "Helvetica", nil
	}
	bold, err := os.ReadFile(fontBold)
	if err != nil {
		return "Helvetica", nil
	}
	pdf.AddUTF8FontFromBytes(fontName, "", regular)
	pdf.AddUTF8FontFromBytes(fontName, "B", bold)
	if pdf.Error() != nil {
		return "", fmt.Errorf("load fonts: %w", pdf.Error())
	}
	return fontName, nil
}

func fitText(pdf *fpdf.Fpdf, family, text string, maxWidth, start, min float64) float64 {
	for size := start; size >= min; size -= 1 {
		pdf.SetFont(family, "B", size)
		if pdf.GetStringWidth(text) <= maxWidth {
			return size
		}
	}
	return min
}
