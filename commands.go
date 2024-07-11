package main

import (
	"fmt"
	//"regexp"
	"sort"
	"strconv"
	"strings"

	tb "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/usher2/u2ckbot/internal/logger"
	pb "github.com/usher2/u2ckbot/msg"
)

const (
	BlockedPattern           = " заблокирован"
	SuffixBlockedPattern     = " блокировки в базовом домене"
	SuffixBlockedPatternPlus = " блокировки в базовом домене+"
	EntryTypeBlockedPattern  = " решения"
)

func botUpdates(c pb.CheckClient, bot *tb.BotAPI, updatesChan tb.UpdatesChannel) {
	for update := range updatesChan {
		switch {
		case update.Message != nil: // ignore any non-Message Updates
			if update.Message.Text != "" {
				if update.Message.Chat.Type == "private" ||
					(update.Message.ReplyToMessage == nil &&
						update.Message.ForwardFromMessageID == 0) {
					var uname string
					// who writing
					if update.Message.From != nil {
						uname = update.Message.From.UserName
					}
					// chat/dialog
					chat := update.Message.Chat
					go Talks(c, bot, uname, chat, "", 0, "", update.Message.Text)
				}
			}
		case update.CallbackQuery != nil:
			var (
				uname string
				req   string
			)
			// who writing
			if update.CallbackQuery.From != nil {
				uname = update.CallbackQuery.From.UserName
			}
			// chat/dialog
			var chat *tb.Chat
			if update.CallbackQuery.Message != nil {
				chat = update.CallbackQuery.Message.Chat
				i := strings.IndexByte(update.CallbackQuery.Message.Text, '\n')
				if i > 0 {
					switch {
					case strings.HasPrefix(update.CallbackQuery.Message.Text[:i], "\U0001f525 ") &&
						strings.HasSuffix(update.CallbackQuery.Message.Text[:i], SuffixBlockedPattern):
						req = "/x " + strings.TrimSuffix(strings.TrimPrefix(update.CallbackQuery.Message.Text[:i], "\U0001f525 "), SuffixBlockedPattern)
					case strings.HasPrefix(update.CallbackQuery.Message.Text[:i], "\U0001f525 ") &&
						strings.HasSuffix(update.CallbackQuery.Message.Text[:i], SuffixBlockedPatternPlus):
						req = "/xx " + strings.TrimSuffix(strings.TrimPrefix(update.CallbackQuery.Message.Text[:i], "\U0001f525 "), SuffixBlockedPatternPlus)
					case strings.HasPrefix(update.CallbackQuery.Message.Text[:i], "\U0001f525 ") &&
						strings.HasSuffix(update.CallbackQuery.Message.Text[:i], BlockedPattern):
						req = strings.TrimSuffix(strings.TrimPrefix(update.CallbackQuery.Message.Text[:i], "\U0001f525 "), BlockedPattern)
					case strings.HasPrefix(update.CallbackQuery.Message.Text[:i], "\U0001f4dc ") &&
						strings.Contains(update.CallbackQuery.Message.Text[:i], "/d_"):
						j1 := strings.Index(update.CallbackQuery.Message.Text[:i], "/d_")
						j2 := strings.IndexByte(update.CallbackQuery.Message.Text[j1:i], ' ')
						if j2 == -1 {
							req = update.CallbackQuery.Message.Text[j1:]
						} else {
							req = update.CallbackQuery.Message.Text[j1 : j1+j2]
						}
					case strings.Contains(update.CallbackQuery.Message.Text[:i], "/n_"):
						j1 := strings.Index(update.CallbackQuery.Message.Text[:i], "/n_")
						j2 := strings.IndexByte(update.CallbackQuery.Message.Text[j1:i], ' ')
						if j2 != -1 {
							req = update.CallbackQuery.Message.Text[j1 : j1+j2]
						} else {
							req = update.CallbackQuery.Message.Text[j1:]
						}
					case strings.Contains(update.CallbackQuery.Message.Text[:i], "/e_"):
						j1 := strings.Index(update.CallbackQuery.Message.Text[:i], "/e_")
						j2 := strings.IndexByte(update.CallbackQuery.Message.Text[j1:i], ' ')
						if j2 != -1 {
							req = update.CallbackQuery.Message.Text[j1 : j1+j2]
						} else {
							req = update.CallbackQuery.Message.Text[j1:]
						}
					case strings.Contains(update.CallbackQuery.Message.Text[:i], "/o_"):
						j1 := strings.Index(update.CallbackQuery.Message.Text[:i], "/o_")
						j2 := strings.IndexByte(update.CallbackQuery.Message.Text[j1:i], ' ')
						if j2 != -1 {
							req = update.CallbackQuery.Message.Text[j1 : j1+j2]
						} else {
							req = update.CallbackQuery.Message.Text[j1:]
						}
					case strings.Contains(update.CallbackQuery.Message.Text[:i], "/wn"):
						j1 := strings.Index(update.CallbackQuery.Message.Text[:i], "/wn")
						j2 := strings.IndexByte(update.CallbackQuery.Message.Text[j1:i], ' ')
						if j2 != -1 {
							req = update.CallbackQuery.Message.Text[j1 : j1+j2]
						} else {
							req = update.CallbackQuery.Message.Text[j1:]
						}
					}
				}
			}

			go bot.Request(tb.NewCallback(update.CallbackQuery.ID, "")) // for some reason

			go Talks(c, bot, uname, chat, "", update.CallbackQuery.Message.MessageID, update.CallbackQuery.Data, req)
		case update.InlineQuery != nil:
			if update.InlineQuery.Query != "" {
				var uname string
				// who writing
				if update.InlineQuery.From != nil {
					uname = update.InlineQuery.From.UserName
				}
				go Talks(c, bot, uname, nil, update.InlineQuery.ID, 0, "", update.InlineQuery.Query)
			}
		}
	}
}

var noAdCount int = 0

const NO_AD_NUMBER = 5

func makePagination(offset TPagination, pages []TPagination) tb.InlineKeyboardMarkup {
	var (
		keyboard [][]tb.InlineKeyboardButton
		o        int
		pict     string
	)
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Tag < pages[j].Tag
	})
	for i := range pages {
		curTag := pages[i].Tag
		if pages[i].Count > PRINT_LIMIT {
			row := tb.NewInlineKeyboardRow()
			if offset.Tag != curTag {
				o = 0
			} else {
				o = offset.Count
			}
			slug := strconv.Itoa(o/PRINT_LIMIT + 1)
			switch curTag {
			case OFFSET_DOMAIN:
				slug = "домен"
			case OFFSET_URL:
				slug = "URL"
			case OFFSET_IP4:
				slug = "IPv4"
			case OFFSET_IP6:
				slug = "IPv6"
			case OFFSET_SUBNET4:
				slug = "подсеть v4"
			case OFFSET_SUBNET6:
				slug = "подсеть v6"
			}
			if pages[i].Count > 2*PRINT_LIMIT {
				pict = "\u23ea"
				if o == 0 {
					pict = "\U000023f9"
				}
				row = append(row,
					tb.NewInlineKeyboardButtonData(fmt.Sprintf("%d  %s", 1, pict),
						fmt.Sprintf("%d:%d", curTag, 0)),
				)
			}
			_o := o - PRINT_LIMIT
			if _o < 0 {
				_o = 0
			}
			pict = "\u23ee"
			if o == 0 {
				pict = "\U000023f9"
			}
			row = append(row,
				tb.NewInlineKeyboardButtonData(fmt.Sprintf("%d  %s", _o/PRINT_LIMIT+1, pict),
					fmt.Sprintf("%d:%d", curTag, _o)),
			)
			row = append(row,
				tb.NewInlineKeyboardButtonData(fmt.Sprintf("\u2022  %s  \u2022", slug),
					fmt.Sprintf("%d:%d", curTag, o)),
			)
			_o = o + PRINT_LIMIT
			if _o > pages[i].Count-(pages[i].Count%PRINT_LIMIT) {
				_o = pages[i].Count - (pages[i].Count % PRINT_LIMIT)
			}
			if _o == pages[i].Count {
				_o -= PRINT_LIMIT
			}
			pict = "\u23ed"
			if o >= _o {
				pict = "\U000023f9"
			}
			_p := _o/PRINT_LIMIT + 1
			row = append(row,
				tb.NewInlineKeyboardButtonData(fmt.Sprintf("%s  %d", pict, _p),
					fmt.Sprintf("%d:%d", curTag, _o)),
			)
			if pages[i].Count > 2*PRINT_LIMIT {
				_o = pages[i].Count - (pages[i].Count % PRINT_LIMIT)
				if _o == pages[i].Count {
					_o -= PRINT_LIMIT
				}
				_p = _o/PRINT_LIMIT + 1
				pict = "\u23e9"
				if o >= _o {
					pict = "\U000023f9"
				}
				row = append(row,
					tb.NewInlineKeyboardButtonData(fmt.Sprintf("%s  %d", pict, _p),
						fmt.Sprintf("%d:%d", curTag, _o)),
				)
			}
			keyboard = append(keyboard, row)
		}
	}
	return tb.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

func sendMessage(bot *tb.BotAPI, chat *tb.Chat, inlineId string, messageId int, text string, offset TPagination, pages []TPagination) {
	if chat != nil {
		if noAdCount >= NO_AD_NUMBER || strings.Contains(text, "Сводная статистика по выгрузке") {
			text += "--- \n" + DonateFooter
			noAdCount = 0
		} else {
			text += Footer
			noAdCount++
		}
		if messageId > 0 {
			msg := tb.NewEditMessageText(chat.ID, messageId, text)
			msg.ParseMode = tb.ModeMarkdown
			msg.DisableWebPagePreview = true
			inlineKeyboard := makePagination(offset, pages)
			if len(inlineKeyboard.InlineKeyboard) > 0 {
				msg.ReplyMarkup = &inlineKeyboard
			}
			_, err := bot.Send(msg)
			if err != nil {
				logger.Warning.Printf("Error sending message: %s\n", err.Error())
			}
		} else {
			msg := tb.NewMessage(chat.ID, text)
			msg.ParseMode = tb.ModeMarkdown
			msg.DisableWebPagePreview = true
			inlineKeyboard := makePagination(offset, pages)
			if len(inlineKeyboard.InlineKeyboard) > 0 {
				msg.ReplyMarkup = inlineKeyboard
			}
			_, err := bot.Send(msg)
			if err != nil {
				logger.Warning.Printf("Error sending message: %s\n", err.Error())
			}
		}
	} else if inlineId != "" {
		article := tb.InlineQueryResultArticle{
			ID:    inlineId,
			Title: "Search result",
			Type:  "article",
			InputMessageContent: tb.InputTextMessageContent{
				Text:                  text + Footer,
				ParseMode:             tb.ModeMarkdown,
				DisableWebPagePreview: true,
			},
		}
		inlineConf := tb.InlineConfig{
			InlineQueryID: inlineId,
			Results:       []interface{}{article},
		}
		if _, err := bot.Request(inlineConf); err != nil {
			logger.Warning.Printf("Error sending answer: %s\n", err.Error())
		}
	}
}

// Handle commands
func Talks(c pb.CheckClient, bot *tb.BotAPI, uname string, chat *tb.Chat, inlineId string, messageId int, callbackData, text string) {
	var (
		reply  string
		pages  []TPagination
		offset TPagination
	)
	if callbackData != "" && strings.IndexByte(callbackData, ':') != -1 {
		i := strings.IndexByte(callbackData, ':')
		if i != len(callbackData)-1 {
			offset.Tag, _ = strconv.Atoi(callbackData[:i])
			offset.Count, _ = strconv.Atoi(callbackData[i+1:])
		}
	}
	// log.Printf("[%s] %d %s", UserName, ChatID, Text)
	if i := strings.IndexByte(text, '\n'); i != -1 {
		text = strings.Trim(text[:i], " ")
	}
	switch {
	case text == "":
		reply = "\U0001f440 Нечего искать\n"
	case text == "/help":
		reply = HelpMessage
	case text == "/helpen":
		reply = HelpMessageEn
	case text == "/donate":
		reply = DonateMessage
	case text == "/start":
		reply = "Приветствую тебя, " + Sanitize(uname) + "!\n"
	case text == "/ping":
		reply = Ping(c)
	case text == "/sum":
		reply = Summarize(c)
	case text == "/wn":
		reply, pages = withoutNoSearch(c, offset)
	case text == "/ck" || text == "/check" || text == "/x":
		reply = HelpMessage
	case strings.HasPrefix(text, "/ck ") || strings.HasPrefix(text, "/check "):
		reply, pages = mainSearch(c, strings.TrimPrefix(strings.TrimPrefix(text, "/ck "), "/check "), offset)
	case strings.HasPrefix(text, "/x "):
		reply, pages = domainSuffixSearch(c, strings.TrimPrefix(text, "/x "), offset, 1)
	case strings.HasPrefix(text, "/xx "):
		reply, pages = domainSuffixSearch(c, strings.TrimPrefix(text, "/xx "), offset, 2)
	case strings.HasPrefix(text, "/n_") || strings.HasPrefix(text, "#"):
		args := ""
		if strings.HasPrefix(text, "/n_") {
			args = strings.TrimPrefix(text, "/n_")
		} else if strings.HasPrefix(text, "#") {
			args = strings.TrimPrefix(text, "#")
		}
		reply, pages = numberSearch(c, args, offset)
	case strings.HasPrefix(text, "/d_") || strings.HasPrefix(text, "&"):
		args := ""
		if strings.HasPrefix(text, "/d_") {
			args = strings.TrimPrefix(text, "/d_")
		} else if strings.HasPrefix(text, "&") {
			args = strings.TrimPrefix(text, "&")
		}
		reply, pages = decisionSearch(c, args, offset)
	case strings.HasPrefix(text, "/e_") || strings.HasPrefix(text, "^"):
		args := ""
		if strings.HasPrefix(text, "/e_") {
			args = strings.TrimPrefix(text, "/e_")
		} else if strings.HasPrefix(text, "^") {
			args = strings.TrimPrefix(text, "^")
		}
		reply, pages = entryTypeSearch(c, args, offset)
	case strings.HasPrefix(text, "/o_") || strings.HasPrefix(text, "!"):
		args := ""
		if strings.HasPrefix(text, "/o_") {
			args = strings.TrimPrefix(text, "/o_")
		} else if strings.HasPrefix(text, "!") {
			args = strings.TrimPrefix(text, "!")
		}
		reply, pages = orgSearch(c, args, offset)
	case strings.HasPrefix(text, "/"):
		reply = "\U0001f523 iNJALID DEJICE\n"
	default:
		reply, pages = mainSearch(c, text, offset)
	}
	if reply == "" {
		reply = "\U0001f463 Ничего не нашлось\n"
	}
	sendMessage(bot, chat, inlineId, messageId, reply, offset, pages)

	// regex, _ := regexp.Compile(`^/([A-Za-z\_\#\&]+)\s*(.*)$`)
	// matches := regex.FindStringSubmatch(text)
	// Debug.Println(pages)
}
