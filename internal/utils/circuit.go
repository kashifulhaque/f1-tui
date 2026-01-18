package utils

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	circuitWidth  = 200
	circuitHeight = 120
)

var asciiCircuits = []string{
	"  ____\n / __ \\\n| |  | |\n| |__| |\n \\____/",
	"  ____\n / __ \\\n| /  \\ |\n| \\__/ |\n \\____/",
	"  _____\n / ___ \\\n| /   \\|\n| \\___/|\n \\_____/",
	"  ____\n / __ \\\n| |  | |\n| |__| |\n \\_  _/",
}

func HasChafa() bool {
	_, err := exec.LookPath("chafa")
	return err == nil
}

func FetchCircuitSVG(wikiURL string) string {
	if wikiURL == "" {
		return ""
	}
	seed := circuitSeed(wikiURL)
	fileName := fmt.Sprintf("f1tui-circuit-%x.svg", seed)
	path := filepath.Join(os.TempDir(), fileName)
	if _, err := os.Stat(path); err == nil {
		return path
	}

	svg := buildCircuitSVG(seed)
	if err := os.WriteFile(path, []byte(svg), 0o600); err != nil {
		return ""
	}
	return path
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

func CircuitASCII(wikiURL string) string {
	if wikiURL == "" {
		return ""
	}
	idx := int(circuitSeed(wikiURL) % uint64(len(asciiCircuits)))
	return asciiCircuits[idx]
}

func RenderCircuitDiagram(wikiURL string) string {
	svg := FetchCircuitSVG(wikiURL)
	if HasChafa() {
		if art := RenderWithChafa(svg); art != "" {
			return art
		}
	}
	return CircuitASCII(wikiURL)
}

func buildCircuitSVG(seed uint64) string {
	rng := rand.New(rand.NewSource(int64(seed)))
	points := make([][2]float64, 0, 8)
	for i := 0; i < 8; i++ {
		angle := float64(i) / 8 * 2 * math.Pi
		rx := 70 + rng.Float64()*30
		ry := 28 + rng.Float64()*18
		jitter := rng.Float64()*6 - 3
		x := float64(circuitWidth)/2 + math.Cos(angle)*rx + jitter
		y := float64(circuitHeight)/2 + math.Sin(angle)*ry + jitter
		points = append(points, [2]float64{x, y})
	}

	var path strings.Builder
	fmt.Fprintf(&path, "M %.1f %.1f", points[0][0], points[0][1])
	for i := 1; i < len(points); i++ {
		fmt.Fprintf(&path, " L %.1f %.1f", points[i][0], points[i][1])
	}
	path.WriteString(" Z")

	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d"><path d="%s" fill="none" stroke="#ffffff" stroke-width="6" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
		circuitWidth,
		circuitHeight,
		circuitWidth,
		circuitHeight,
		path.String(),
	)
}

func circuitSeed(input string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(input))
	return h.Sum64()
}
