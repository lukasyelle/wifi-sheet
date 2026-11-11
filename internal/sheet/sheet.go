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
	corner   = 0.20
	subtitle = "Scan to join this WiFi"
	fontName = "ssid"
	fontReg  = "/System/Library/Fonts/Supplemental/Arial.ttf"
	fontBold = "/System/Library/Fonts/Supplemental/Arial Bold.ttf"
)

// Network is one WiFi entry rendered on a sheet.
type Network struct {
	SSID  string
	QRPNG []byte
}

// WriteFile renders a single US Letter PDF with a nature background and one
// card per network (SSID + scannable WiFi QR). Passwords are not printed.
func WriteFile(path string, networks []Network, background []byte) error {
	if len(networks) == 0 {
		return fmt.Errorf("at least one network is required")
	}
	for i, n := range networks {
		if n.SSID == "" {
			return fmt.Errorf("network %d: ssid is required", i+1)
		}
		if len(n.QRPNG) == 0 {
			return fmt.Errorf("network %d (%s): qr image is required", i+1, n.SSID)
		}
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

	if len(networks) == 1 {
		drawSingleCard(pdf, family, networks[0])
	} else {
		drawGrid(pdf, family, networks)
	}

	if pdf.Error() != nil {
		return fmt.Errorf("render pdf: %w", pdf.Error())
	}
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

func drawSingleCard(pdf *fpdf.Fpdf, family string, n Network) {
	const (
		cardW  = 5.20
		cardH  = 6.20
		qrSize = 3.55
	)
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
	titleSize := fitText(pdf, family, n.SSID, titleMaxW, 26, 13)
	pdf.SetFont(family, "B", titleSize)
	pdf.SetTextColor(28, 32, 36)
	pdf.SetXY(cardX, cardY+0.38)
	pdf.CellFormat(cardW, 0.42, n.SSID, "", 0, "C", false, 0, "")

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
	pdf.RegisterImageOptionsReader("qr-0", qrOpt, bytes.NewReader(n.QRPNG))
	qrX := cardX + (cardW-qrSize)/2
	qrY := cardY + 1.42
	pdf.ImageOptions("qr-0", qrX, qrY, qrSize, qrSize, false, qrOpt, 0, "")
}

func drawGrid(pdf *fpdf.Fpdf, family string, networks []Network) {
	cols, rows := gridDims(len(networks))
	const (
		margin = 0.42
		gap    = 0.22
	)
	usableW := pageW - 2*margin
	usableH := pageH - 2*margin
	cardW := (usableW - float64(cols-1)*gap) / float64(cols)
	cardH := (usableH - float64(rows-1)*gap) / float64(rows)

	// Cap tall single-row cards so two side-by-side networks stay compact.
	maxCardH := cardW * 1.45
	if maxCardH < 3.4 {
		maxCardH = 3.4
	}
	if maxCardH > 6.2 {
		maxCardH = 6.2
	}
	if cardH > maxCardH {
		cardH = maxCardH
	}

	gridW := float64(cols)*cardW + float64(cols-1)*gap
	gridH := float64(rows)*cardH + float64(rows-1)*gap
	originX := (pageW - gridW) / 2
	originY := (pageH - gridH) / 2

	// Keep QR dominant inside each card while leaving room for title + subtitle.
	titleBand := 0.95
	if cardH < 2.8 {
		titleBand = 0.78
	}
	sidePad := 0.22
	qrSize := cardW - 2*sidePad
	if maxQR := cardH - titleBand - 0.18; qrSize > maxQR {
		qrSize = maxQR
	}
	if qrSize < 1.05 {
		qrSize = 1.05
	}

	titleStart := 18.0
	titleMin := 10.0
	subSize := 10.0
	sub := "Scan to join"
	if cols == 1 {
		titleStart = 22
		titleMin = 12
		subSize = 11
		sub = subtitle
	}

	for i, n := range networks {
		col := i % cols
		row := i / cols
		x := originX + float64(col)*(cardW+gap)
		y := originY + float64(row)*(cardH+gap)
		drawCard(pdf, family, x, y, cardW, cardH, qrSize, n, fmt.Sprintf("qr-%d", i), sub, titleStart, titleMin, subSize)
	}
}

func gridDims(n int) (cols, rows int) {
	switch {
	case n <= 1:
		return 1, 1
	case n == 2:
		return 2, 1
	case n <= 4:
		return 2, 2
	case n <= 6:
		return 2, 3
	case n <= 9:
		return 3, 3
	default:
		cols = 3
		rows = (n + cols - 1) / cols
		return cols, rows
	}
}

func drawCard(
	pdf *fpdf.Fpdf,
	family string,
	cardX, cardY, cardW, cardH, qrSize float64,
	n Network,
	qrName, sub string,
	titleStart, titleMin, subSize float64,
) {
	pdf.SetAlpha(0.16, "Normal")
	pdf.SetFillColor(18, 22, 28)
	pdf.RoundedRect(cardX+0.07, cardY+0.09, cardW, cardH, corner, "1234", "F")

	pdf.SetAlpha(0.94, "Normal")
	pdf.SetFillColor(255, 255, 255)
	pdf.RoundedRect(cardX, cardY, cardW, cardH, corner, "1234", "F")
	pdf.SetAlpha(1, "Normal")

	titleMaxW := cardW - 0.40
	titleSize := fitText(pdf, family, n.SSID, titleMaxW, titleStart, titleMin)
	pdf.SetFont(family, "B", titleSize)
	pdf.SetTextColor(28, 32, 36)
	pdf.SetXY(cardX, cardY+0.22)
	pdf.CellFormat(cardW, 0.36, n.SSID, "", 0, "C", false, 0, "")

	pdf.SetFont(family, "", subSize)
	pdf.SetTextColor(92, 98, 106)
	pdf.SetXY(cardX, cardY+0.58)
	pdf.CellFormat(cardW, 0.24, sub, "", 0, "C", false, 0, "")

	ruleY := cardY + 0.88
	pdf.SetDrawColor(220, 223, 226)
	pdf.SetLineWidth(0.008)
	inset := cardW * 0.22
	if inset < 0.35 {
		inset = 0.35
	}
	if inset > 1.15 {
		inset = 1.15
	}
	pdf.Line(cardX+inset, ruleY, cardX+cardW-inset, ruleY)

	qrOpt := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
	pdf.RegisterImageOptionsReader(qrName, qrOpt, bytes.NewReader(n.QRPNG))
	qrX := cardX + (cardW-qrSize)/2
	qrY := cardY + (cardH-qrSize)/2 + 0.28
	if qrY+qrSize > cardY+cardH-0.16 {
		qrY = cardY + cardH - qrSize - 0.16
	}
	pdf.ImageOptions(qrName, qrX, qrY, qrSize, qrSize, false, qrOpt, 0, "")
}

// Filename returns a filesystem-safe PDF name for one or more SSIDs.
func Filename(ssids ...string) string {
	if len(ssids) == 0 {
		return "wifi-networks.pdf"
	}
	parts := make([]string, 0, len(ssids))
	for _, ssid := range ssids {
		part := sanitizeName(ssid)
		if part == "" {
			part = "network"
		}
		parts = append(parts, part)
	}
	joined := strings.Join(parts, "-")
	const maxLen = 80
	if len(joined) > maxLen {
		joined = strings.Trim(joined[:maxLen], "-")
		if joined == "" {
			joined = "networks"
		}
	}
	return "wifi-" + joined + ".pdf"
}

func sanitizeName(ssid string) string {
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
	return strings.Trim(b.String(), "-")
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
