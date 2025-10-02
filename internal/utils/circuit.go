package utils

import (
	"os/exec"
)

// HasChafa checks if chafa is available on the system
func HasChafa() bool {
	_, err := exec.LookPath("chafa")
	return err == nil
}

// FetchCircuitSVG attempts to fetch circuit SVG (placeholder implementation)
func FetchCircuitSVG(wikiURL string) string {
	// Placeholder: In a real implementation, this would fetch and parse SVG
	return ""
}

// RenderWithChafa renders an image using chafa terminal graphics
func RenderWithChafa(svgPath string) string {
	if svgPath == "" {
		return ""
	}
	cmd := exec.Command("chafa", svgPath, "--size=40x12")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}
