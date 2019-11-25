package main

const (
	HelpMessage string = "*Commands:*\n" +
		"/help - This message\n\n" +
		"*Simple usage*\n\n" +
		"Send an Ip-address or Domain or URL to the bot for restriction checking \n\n" +
		"Send a record ID (ussualy started with #) to the bot for details\n\n" +
		"*Legend*\n\n" +
		"\U000026d4 URL blocking type. It's default blockig type. Providers MUST block ONLY plain HTTP traffic with certain URL\n" +
		"\U0001f4db HTTPS blocking type. It's not standart blocking type. It's URL blocking type but with HTTPS scheme. Providers MUST block domain name from URL by SNI or by DNS interception\n" +
		"\U0001f6ab Domain blockig type. Providers MUST block plain HTTP traffic with certain Host header and domain name by SNI or by DNS interception\n" +
		"\U0001f506 Wildcard blockig type. It's similar to domain blocking type but providers MUST block domain name with wildcard name\n" +
		"\u274c IP clockuing type. Providers MUST block whole IP-address or subnet\n"

	Footer string = "\n--- \n" +
		"https://t.me/usher2 project\nhttps://www.paypal.me/schors\nETH: 0xba53cebd99157bf412a6bb91165e7dff29abd0a2\nWMZ: Z991867115444\n"
)
