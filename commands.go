//go:generate protoc -I msg --go_out=plugins=grpc:msg msg/msg.proto
package main

import (
	"regexp"

	tb "github.com/go-telegram-bot-api/telegram-bot-api"

	pb "github.com/usher2/u2ckbot/msg"
)

// Handle Chat message
func Talks(c pb.CheckClient, bot *tb.BotAPI, update tb.Update) {
	// who writing
	UserName := ""
	if update.Message.From != nil {
		UserName = update.Message.From.UserName
	}
	// ID of chat/dialog
	// maiby eq UserID or public chat
	ChatID := update.Message.Chat.ID
	bot.Send(tb.NewChatAction(ChatID, "typing"))
	// message text
	Text := update.Message.Text
	//log.Printf("[%s] %d %s", UserName, ChatID, Text)
	regex, _ := regexp.Compile(`^/([A-Za-z\_]+)\s*(.*)$`)
	matches := regex.FindStringSubmatch(Text)
	// hanlde chat commands
	if len(matches) > 0 {
		var reply string
		comm := matches[1]
		commArgs := []string{""}
		if len(matches) >= 3 {
			commArgs = regexp.MustCompile(`\s+`).Split(matches[2], -1)
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
				reply = mainSearch(c, commArgs[len(commArgs)-1])
			} else {
				reply = "😱 Noting to search\n"
			}
		case `start`:
			reply = "Glad to see you, " + UserName + "!\n"
		//case `ping`:
		//	reply = Ping(c)
		default:
			reply = "😱 Unknown command\n"
		}
		if reply != `` {
			msg := tb.NewMessage(ChatID, reply+Footer)
			msg.ParseMode = tb.ModeMarkdown
			msg.DisableWebPagePreview = true
			_, err := bot.Send(msg)
			if err != nil {
				Warning.Printf("Error sending message: %s\n", err.Error())
			}
		}
	} else {
		msg := tb.NewMessage(ChatID, mainSearch(c, Text)+Footer)
		msg.ParseMode = tb.ModeMarkdown
		msg.DisableWebPagePreview = true
		_, err := bot.Send(msg)
		if err != nil {
			Warning.Printf("Error sending message: %s\n", err.Error())
		}
	}
}
