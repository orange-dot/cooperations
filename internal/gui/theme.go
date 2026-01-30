package gui

import (
	"image/color"
	"strconv"
	"strings"
)

// Theme defines the color palette used by the GUI.
type Theme struct {
	Background    color.NRGBA
	PanelBg       color.NRGBA
	Border        color.NRGBA
	BorderActive  color.NRGBA
	TextPrimary   color.NRGBA
	TextSecondary color.NRGBA
	Success       color.NRGBA
	Error         color.NRGBA
	Warning       color.NRGBA
	Accent        color.NRGBA
	Cyan          color.NRGBA
}

// DefaultTheme is the futuristic dark theme palette.
var DefaultTheme = Theme{
	Background:    HexToNRGBA("#0a0e17"),
	PanelBg:       HexToNRGBA("#0d1520"),
	Border:        HexToNRGBA("#1a3a4a"),
	BorderActive:  HexToNRGBA("#00ffff"),
	TextPrimary:   HexToNRGBA("#ffffff"),
	TextSecondary: HexToNRGBA("#8899aa"),
	Success:       HexToNRGBA("#00ff88"),
	Error:         HexToNRGBA("#ff4466"),
	Warning:       HexToNRGBA("#ffaa00"),
	Accent:        HexToNRGBA("#ff00ff"),
	Cyan:          HexToNRGBA("#00ffff"),
}

// HexToNRGBA converts a hex color string into color.NRGBA.
// Accepts forms: "#RRGGBB", "RRGGBB", "#RRGGBBAA", "RRGGBBAA".
func HexToNRGBA(hex string) color.NRGBA {
	s := strings.TrimSpace(hex)
	if strings.HasPrefix(s, "#") {
		s = s[1:]
	}
	if len(s) != 6 && len(s) != 8 {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}

	parseByte := func(part string) (uint8, bool) {
		v, err := strconv.ParseUint(part, 16, 8)
		if err != nil {
			return 0, false
		}
		return uint8(v), true
	}

	r, ok := parseByte(s[0:2])
	if !ok {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}
	g, ok := parseByte(s[2:4])
	if !ok {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}
	b, ok := parseByte(s[4:6])
	if !ok {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}

	a := uint8(255)
	if len(s) == 8 {
		aa, ok := parseByte(s[6:8])
		if !ok {
			return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
		}
		a = aa
	}

	return color.NRGBA{R: r, G: g, B: b, A: a}
}

// WithAlpha returns the same color with the provided alpha.
func WithAlpha(c color.NRGBA, alpha uint8) color.NRGBA {
	c.A = alpha
	return c
}