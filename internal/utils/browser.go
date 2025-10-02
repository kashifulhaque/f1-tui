package utils

import "github.com/pkg/browser"

// OpenURL opens a URL in the default browser
func OpenURL(url string) {
	_ = browser.OpenURL(url)
}
