package pdf

import (
	"fmt"
	"strings"
)

// PageSize describes a printable page in the units Chrome DevTools expects
// (inches). Use the predefined constants A4 / Letter or ParsePageSize.
type PageSize struct {
	Name   string
	Width  float64 // inches
	Height float64 // inches
}

var (
	A4     = PageSize{Name: "A4", Width: 8.27, Height: 11.69}      // 210 × 297 mm
	Letter = PageSize{Name: "Letter", Width: 8.5, Height: 11.0}    // 8.5 × 11 in
)

// ParsePageSize accepts case-insensitive "A4" or "Letter".
func ParsePageSize(s string) (PageSize, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "a4":
		return A4, nil
	case "letter":
		return Letter, nil
	default:
		return PageSize{}, fmt.Errorf("tamaño de página no soportado: %q (válidos: A4, Letter)", s)
	}
}
