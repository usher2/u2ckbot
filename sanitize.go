package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	// res = strings.ReplaceAll(s, "_", "\\_")
	// res = strings.ReplaceAll(res, "*", "\\*")
	// res = strings.ReplaceAll(res, "`", "\\`")
	return tgbotapi.EscapeText(tgbotapi.ModeMarkdown, s)
}
