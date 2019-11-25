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

	DonateMessage string = "☀️ *Пожертвования по подписке:*\n" +
		"https://www.patreon.com/usher2\n\n" +
		"⭐️ *Традиционные способы:*" +
		"PayPal: https://www.paypal.me/schors\n" +
		"Яндекс.Деньги: http://yasobe.ru/na/schors\n" +
		"WMP: P603777732896\n" +
		"WMZ: Z991867115444\n" +
		"WME: E261636674470\n" +
		"WMX: X862559021665\n\n" +
		"🏵 *Сделать меня криптомагнатом:*\n" +
		"BTC: 18YFeAV12ktBxv9hy4wSiSCUXXAh5VR7gE\n" +
		"LTC: LVXP51M8MrzaEQi6eBEGWpTSwckybqHU5s\n" +
		"ETH: 0xba53cebd99157bf412a6bb91165e7dff29abd0a2\n" +
		"ZEC: t1McmUhzdsauoXpiu2yCjNpnLKGGH225aAW\n" +
		"DGE: D8cZwBsVp1hW4mjTCgspEKG5TpPZycTJBn\n" +
		"BCH: 1FiXmPZ6eecHVaZbgdadAuzQLU9kqdSzVN\n" +
		"ETC: 0xeb990a29d4f870b5fdbe331db90d9849ce3dae77\n" +
		"WAX: 0xba53cebd99157bf412a6bb91165e7dff29abd0a2\n\n" +
		"✈️ *Бонусные программы:*\n" +
		"Аэрофлот-бонус: 1045433852\n" +
		"S7-бонус: 929102200\n\n" +
		"🍭 Мой вишлист: http://mywishlist.ru/me/schors"

	Footer string = "\n--- \n" +
		"https://t.me/usher2 project\nhttps://www.paypal.me/schors\nETH: 0xba53cebd99157bf412a6bb91165e7dff29abd0a2\nWMZ: Z991867115444\n"
)
