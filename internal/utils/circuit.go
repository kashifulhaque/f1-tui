package utils

import (
	"os/exec"
)

func HasChafa() bool {
	_, err := exec.LookPath("chafa")
	return err == nil
}

func FetchCircuitSVG(wikiURL string) string {
	return ""
}

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
