package utils

import "github.com/pkg/browser"

func OpenURL(url string) {
	_ = browser.OpenURL(url)
}
