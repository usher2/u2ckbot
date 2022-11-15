package main

const (
	HelpMessageEn string = "*Commands:*\n" +
		"/helpen - This message\n" +
		"/help - руководство на русском\n" +
		"/ck <website or IP>\n/check <website or IP> - check something - IP-address or Domain or URL\n" +
		"/donate - Information about donation methods\n\n" +
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
		"/ck <САЙТ или IP-адрес>\n/check <САЙТ или IP-адрес> - проверить IP адрес, домен или URL\n" +
		"/donate - получить информацию о способах пожертвований\n\n" +
		"*Основные возможности*\n\n" +
		"Отправьте IP-адрес или домен, или URL боту для проверки наличия их в списке блокировок\n\n" +
		"Отправьте ID записи (обозначена '#' в ответах) боту для получения подробностей\n\n" +
		"Или воспользуйтесь командами /ck или /check в группах\n\n" +
		"/n\\_<НОМЕР> - показывает подробную информацию о данном номере записи в реестре\n\n" +
		"/d\\_<ИДЕНТИФИКАТОР> - показывает список записей в реестре по данному решению\n\n" +
		"*Обозначения*\n\n" +
		"\U000026d4 блокировка по URL. Тип блокировки по умолчанию. Провайдеры ОБЯЗАНЫ фильтровать ТОЛЬКО простой HTTP трафик на определённые URL\n" +
		"\U0001f4db блокировка HTTPS. Неформальный тип блокировк. Это блокировка по URL с HTTPS схемой. Провайдеры ОБЯЗАНЫ блокировать трафик на домен из URL путем чтения SNI или путём перехвата DNS запросов\n" +
		"\U0001f6ab блокировка по домену. Провайдеры ОБЯЗАНЫ фильтровать простой HTTP трафик с определенным заголовком Host и трафик на домен путем чтения SNI или путём перехвата DNS запросов\n" +
		"\U0001f506 блокировка по маске. Тоже самое, что блокировка по домену, только провайдеры ОБЯЗАНЫ сопоставлять доменное имя с шаблоном\n" +
		"\u274c блокировка по IP адресу. Провайдеры ОБЯЗАНЫ блокировать ВЕСЬ IP адрес или подсеть\n"

	DonateMessage string = "\U00002b50 *Традиционные способы:*\n" +
		"Paypal: https://www.paypal.me/schorsx\n" +
		"ЮMoney (бывшие ЯДеньги): https://sobe.ru/na/m2i2s077M0g2\n" +
		"\U0001f3f5 *Сделать меня криптомагнатом:*\n" +
		"BTC: `18YFeAV12ktBxv9hy4wSiSCUXXAh5VR7gE`\n" +
		"LTC: `LVXP51M8MrzaEQi6eBEGWpTSwckybqHU5s`\n" +
		"ETH: `0xba53cebd99157bf412a6bb91165e7dff29abd0a2`\n" +
		"ZEC: `t1McmUhzdsauoXpiu2yCjNpnLKGGH225aAW`\n" +
		"BCH: `1FiXmPZ6eecHVaZbgdadAuzQLU9kqdSzVN`\n" +
		"ETC: `0xeb990a29d4f870b5fdbe331db90d9849ce3dae77`\n" +
		"TON: `EQBrl8BNLWNVvmSCZDNexzoGQLIojnp4xNDT6Wf4AFX4S\\_57`\n\n"
		//"\U0001f36d Мой вишлист: http://mywishlist.ru/me/schors\n\n"

	Footer string = "" /*"--- \n" +
	"Часть проекта @usher2\n\n" /*+
	"\U000026a0 Я хочу тонко намекнуть на толстые обстоятельства. " +
	"Сейчас весь краудфайндинг в 0 уходит на оплату " +
	"хостинга проектов Эшер II - сайт, сбор выгрузок, бот этот. " +
	"Даже писать посты стало решительно некогда. /donate \n" */

	DonateFooter string = "Хочу новый ноут и мониторы: /donate\n"
)
