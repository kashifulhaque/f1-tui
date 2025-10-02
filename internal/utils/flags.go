package utils

import "strings"

func CountryCodeToFlag(code string) string {
	if len(code) != 2 {
		return ""
	}
	code = strings.ToUpper(code)

	r1 := rune(code[0]) - 'A' + 0x1F1E6
	r2 := rune(code[1]) - 'A' + 0x1F1E6

	return string([]rune{r1, r2})
}

func CountryNameToCode(name, gpName, circuitName string) string {
	m := map[string]string{
		"Australia":            "AU",
		"Austria":              "AT",
		"Bahrain":              "BH",
		"Belgium":              "BE",
		"Brazil":               "BR",
		"Canada":               "CA",
		"China":                "CN",
		"United Arab Emirates": "AE", // Abu Dhabi
		"France":               "FR",
		"Germany":              "DE",
		"Hungary":              "HU",
		"Italy":                "IT",
		"Japan":                "JP",
		"Mexico":               "MX",
		"Monaco":               "MC",
		"Netherlands":          "NL",
		"Azerbaijan":           "AZ",
		"Qatar":                "QA",
		"Saudi Arabia":         "SA",
		"Singapore":            "SG",
		"South Korea":          "KR",
		"Spain":                "ES",
		"United Kingdom":       "GB",
		"United States":        "US",
	}

	if code, ok := m[name]; ok {
		return code
	}

	switch {
	case strings.Contains(gpName, "Miami"):
		return "US"
	case strings.Contains(gpName, "Las Vegas"):
		return "US"
	case strings.Contains(gpName, "British"):
		return "GB"
	case strings.Contains(gpName, "Abu Dhabi"):
		return "AE"
	case strings.Contains(gpName, "Baku"):  // Azerbaijan GP
		return "AZ"
	case strings.Contains(gpName, "United States"):
		return "US"
	}

	return ""
}
