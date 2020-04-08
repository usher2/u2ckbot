package main

import (
	"strings"

	"golang.org/x/net/idna"
)

func PrintedDomain(domain string) string {
	_domain, err := idna.ToUnicode(domain)
	if err != nil {
		return domain
	}
	return _domain
}

func Sanitize(s string) (res string) {
	res = strings.ReplaceAll(s, "_", "\\_")
	res = strings.ReplaceAll(res, "*", "\\*")
	res = strings.ReplaceAll(res, "`", "\\`")
	return
}
