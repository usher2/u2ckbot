package main

const (
	HelpMessageEn string = "*Commands:*\n" +
		"/helpen - This message\n" +
		"/help - руководство на русском\n" +
		"/ck <website or IP>\n/check <website or IP> - check something - IP-address or Domain or URL\n\n" +
		"*Simple usage*\n\n" +
		"Send an IP-address or Domain or URL to the bot for restriction checking \n\n" +
		"Send a record ID (ussualy started with #) to the bot for details\n\n" +
		"Or use /ck or /check commands in group chats\n\n" +
		"*Legend*\n\n" +
		"\U000026d4 URL blocking type. It's default blockig type. Providers MUST block ONLY plain HTTP traffic with certain URL\n" +
		"\U0001f4db HTTPS blocking type. It's not standart blocking type. It's URL blocking type but with HTTPS scheme. Providers MUST block domain name from URL by SNI or by DNS interception\n" +
		"\U0001f6ab Domain blockig type. Providers MUST block plain HTTP traffic with certain Host header and domain name by SNI or by DNS interception\n" +
		"\U0001f506 Wildcard blockig type. It's similar to domain blocking type but providers MUST block domain name with wildcard name\n" +
		"\u274c IP blocking type. Providers MUST block whole IP-address or subnet\n"

	HelpMessage string = "*Команды:*\n" +
		"/helpen - manual in english\n" +
		"/help - это сообщение\n" +
		"/ck <САЙТ или IP-адрес>\n/check <САЙТ или IP-адрес> - проверить IP адрес, домен или URL\n\n" +
		"*Основные возможности*\n\n" +
		"Отправьте IP-адрес или домен, или URL боту для проверки наличия их в списке блокировок\n\n" +
		"Отправьте ID записи (обозначена '#' в ответах) боту для получения подробностей\n" +
		"Или воспользуйтесь командами /ck или /check в группах\n\n" +
		"*Обозначения*\n\n" +
		"\U000026d4 блокировка по URL. Тип блокировки по умолчанию. Провайдеры ОБЯЗАНЫ фильтровать ТОЛЬКО простой HTTP трафик на определённые URL\n" +
		"\U0001f4db блокировка HTTPS. Неформальный тип блокировк. Это блокировка по URL с HTTPS схемой. Провайдеры ОБЯЗАНЫ блокировать трафик на домен из URL путем чтения SNI или путём перехвата DNS запросов\n" +
		"\U0001f6ab блокировка по домену. Провайдеры ОБЯЗАНЫ фильтровать простой HTTP трафик с определенным заголовком Host и трафик на домен путем чтения SNI или путём перехвата DNS запросов\n" +
		"\U0001f506 блокировка по маске. Тоже самое, чо блокировка по домену, только провайдеры ОБЯЗАНЫ сопостовлять доменное имя с шаблоном\n" +
		"\u274c блокировка по IP адресу. Провайдеры ОБЯЗАНЫ блокировать ВЕСЬ IP адрес или подсеть\n"

	DonateMessage string = "\u2600 *Пожертвования по подписке:*\n" +
		"https://www.patreon.com/usher2\n\n" +
		"\U00002b50 *Традиционные способы:*\n" +
		"PayPal: https://www.paypal.me/schors\n" +
		"Яндекс.Деньги: http://yasobe.ru/na/schors\n" +
		"WMP: P603777732896\n" +
		"WMZ: Z991867115444\n" +
		"WME: E261636674470\n" +
		"WMX: X862559021665\n\n" +
		"\U0001f3f5 *Сделать меня криптомагнатом:*\n" +
		"BTC: 18YFeAV12ktBxv9hy4wSiSCUXXAh5VR7gE\n" +
		"LTC: LVXP51M8MrzaEQi6eBEGWpTSwckybqHU5s\n" +
		"ETH: 0xba53cebd99157bf412a6bb91165e7dff29abd0a2\n" +
		"ZEC: t1McmUhzdsauoXpiu2yCjNpnLKGGH225aAW\n" +
		"DGE: D8cZwBsVp1hW4mjTCgspEKG5TpPZycTJBn\n" +
		"BCH: 1FiXmPZ6eecHVaZbgdadAuzQLU9kqdSzVN\n" +
		"ETC: 0xeb990a29d4f870b5fdbe331db90d9849ce3dae77\n" +
		"WAX: 0xba53cebd99157bf412a6bb91165e7dff29abd0a2\n\n" +
		"\U00002708 *Бонусные программы:*\n" +
		"Аэрофлот-бонус: 1045433852\n" +
		"S7-бонус: 929102200\n\n" +
		"\U0001f36d Мой вишлист: http://mywishlist.ru/me/schors\n\n" +
		"\U00002708 *Игрушки:*\n" +
		"World of Warships: Phil\\_Kulin\n"

	Footer string = "--- \n" +
		"Часть проекта @usher2\n\n" /*+
		"\U000026a0 Я хочу тонко намекнуть на толстые обстоятельства. " +
		"Сейчас весь краудфайндинг в 0 уходит на оплату " +
		"хостинга проектов Эшер II - сайт, сбор выгрузок, бот этот. " +
		"Даже писать посты стало решительно некогда. /donate \n" */

	DonateFooter string = "Для дальнейшей поддержки и разработки бота и его окружения " +
		"требуется финансовая поддержка. /donate\n"
)
