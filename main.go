//go:generate protoc -I msg --go_out=plugins=grpc:msg msg/msg.proto
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	tb "github.com/go-telegram-bot-api/telegram-bot-api"

	pb "github.com/usher-2/u2ckbot/msg"
	"google.golang.org/grpc"
)

type TypeConfig struct {
	// Config
	Token         string
	UserFile      string
	CkDumpService string
}

// connect to Telegram API
func GetBot(token string) *tb.BotAPI {
	bot, err := tb.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return bot
}

// initialize update chan
func GetUpdatesChan(bot *tb.BotAPI) <-chan tb.Update {
	c := tb.NewUpdate(0)
	c.Timeout = 60
	updates, err := bot.GetUpdatesChan(c)
	if err != nil {
		log.Panic(err)
	}
	return updates
}

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
		case `check` || `ck`:
			if len(commArgs) > 0 {
				reply = mainSearch(c, commArgs[0])
			} else {
				reply = "😱 Noting to search\n"
			}
		case `start`:
			reply = "Glad to see you, " + UserName + "!"
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

func main() {
	confFilename := flag.String("c", "u2ckbot.conf", "Configuration file")
	flag.Parse()
	config, err := ReadConfigFile(*confFilename)
	if err != nil {
		log.Fatal(err)
	}
	logLevel := config.GetString("LogLevel", "Debug")
	if logLevel == "Info" {
		logInit(ioutil.Discard, os.Stdout, os.Stderr, os.Stderr)
	} else if logLevel == "Warning" {
		logInit(ioutil.Discard, ioutil.Discard, os.Stderr, os.Stderr)
	} else if logLevel == "Error" {
		logInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	} else {
		logInit(os.Stderr, os.Stdout, os.Stderr, os.Stderr)
	}
	//gRPC
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	//opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(config.GetString("CkDumpServer", "localhost:50001"), opts...)
	if err != nil {
		fmt.Printf("fail to dial: %v", err)
	}
	defer conn.Close()
	fmt.Printf("Connect...\n")
	c := pb.NewCheckClient(conn)
	// connect to Telegram API
	Bot := GetBot(config.GetString("Token", ""))
	// init Users cache
	//ReadUsers(config.UserFile)
	// init update chan
	Updates := GetUpdatesChan(Bot)
	// read updates
	for {
		select {
		case update := <-Updates:
			if update.Message != nil { // ignore any non-Message Updates
				go Talks(c, Bot, update)
			}
		}
	}
}
