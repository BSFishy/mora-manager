package main

import (
	"regexp"
	"strings"
)

func Assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

// regex that matches anything NOT a lowercase letter, digit, "-" or "."
var (
	invalidSubdomainChars = regexp.MustCompile(`[^a-z0-9\-\.]`)
	invalidLabelChars     = regexp.MustCompile(`[^a-z0-9\-]`)
)

// sanitizeDNS1123Subdomain applies lowercase, replaces invalid chars with "-",
// trims leading/trailing non-alphanumeric, and enforces max length (253).
func sanitizeDNS1123Subdomain(name string) string {
	name = strings.ToLower(name)
	name = invalidSubdomainChars.ReplaceAllString(name, "-")
	// Trim any leading or trailing '-' or '.' so it starts/ends with [a-z0-9]
	name = strings.Trim(name, "-.")
	if len(name) > 253 {
		name = name[:253]
	}
	return name
}

// sanitizeDNS1123Label is similar but only allows [a-z0-9\-], max 63 chars
func sanitizeDNS1123Label(name string) string {
	name = strings.ToLower(name)
	name = invalidLabelChars.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if len(name) > 63 {
		name = name[:63]
	}
	return name
}
