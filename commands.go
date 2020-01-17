//go:generate protoc -I msg --go_out=plugins=grpc:msg msg/msg.proto
package main

import (
	"regexp"
	"strings"

	tb "github.com/go-telegram-bot-api/telegram-bot-api"

	pb "github.com/usher2/u2ckbot/msg"
)

func botUpdates(c pb.CheckClient, bot *tb.BotAPI, updatesChan tb.UpdatesChannel) {
	for {
		select {
		case update := <-updatesChan:
			if update.Message != nil { // ignore any non-Message Updates
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
						go Talks(c, bot, uname, chat, "", update.Message.Text)
					}
				}
			} else if update.InlineQuery != nil {
				if update.InlineQuery.Query != "" {
					var uname string
					// who writing
					if update.InlineQuery.From != nil {
						uname = update.InlineQuery.From.UserName
					}
					go Talks(c, bot, uname, nil, update.InlineQuery.ID, update.InlineQuery.Query)
				}
			}
		}
	}
}

var noAdCount int = 0

const NO_AD_NUMBER = 20

func sendMessage(bot *tb.BotAPI, chat *tb.Chat, inlineId string, text string) {
	if chat != nil {
		if noAdCount >= NO_AD_NUMBER {
			text += Footer + DonateFooter
			noAdCount = 0
		} else {
			text += Footer
			noAdCount += 1
		}
		msg := tb.NewMessage(chat.ID, text)
		msg.ParseMode = tb.ModeMarkdown
		msg.DisableWebPagePreview = true
		_, err := bot.Send(msg)
		if err != nil {
			Warning.Printf("Error sending message: %s\n", err.Error())
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
		if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
			Warning.Printf("Error sending answer: %s\n", err.Error())
		}
	}
}

// Handle commands
func Talks(c pb.CheckClient, bot *tb.BotAPI, uname string, chat *tb.Chat, inlineId string, text string) {
	var reply string
	if chat != nil {
		bot.Send(tb.NewChatAction(chat.ID, "typing"))
	}
	//log.Printf("[%s] %d %s", UserName, ChatID, Text)
	regex, _ := regexp.Compile(`^/([A-Za-z\_]+)\s*(.*)$`)
	matches := regex.FindStringSubmatch(text)
	// hanlde chat commands
	if len(matches) > 0 {
		comm := matches[1]
		commArgs := []string{""}
		if len(matches) >= 3 {
			commArgs = regexp.MustCompile(`\s+`).Split(matches[2], -1)
			if bot.Self.UserName != "" {
				for i, s := range commArgs {
					commArgs[i] = strings.TrimSuffix(s, "@"+bot.Self.UserName)
				}
			}
		}
		switch comm {
		case `help`:
			reply = HelpMessage
		case `helpru`:
			reply = HelpMessageRu
		case `donate`:
			reply = DonateMessage
		case `n_`, `ck`, `check`:
			if len(commArgs) > 0 {
				reply = mainSearch(c, commArgs[0])
			} else {
				reply = "😱 Noting to search\n"
			}
		case `start`:
			reply = "Glad to see you, " + uname + "!\n"
			//case `ping`:
			//	reply = Ping(c)
			//default:
			//	reply = "😱 Unknown command\n"
		}
		if reply != "" {
			sendMessage(bot, chat, inlineId, reply)
		}
	} else {
		if text[0] != '/' {
			reply = mainSearch(c, text)
			sendMessage(bot, chat, inlineId, reply)
		}
	}
}
