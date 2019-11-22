package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	tb "github.com/go-telegram-bot-api/telegram-bot-api"
)

var HelpMessage string = `Commands:
/help - This message
/start - Add telegram user to the bot roster
`

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
func Talks(usersFileName string, bot *tb.BotAPI, update tb.Update) {
	// who writing
	UserName := update.Message.From.UserName
	// ID of chat/dialog
	// maiby eq UserID or public chat
	ChatID := update.Message.Chat.ID
	// message text
	Text := update.Message.Text
	//log.Printf("[%s] %d %s", UserName, ChatID, Text)
	regex, _ := regexp.Compile(`^/([A-Za-z\_]+)\s*(.*)$`)
	matches := regex.FindStringSubmatch(Text)
	// hanlde chat commands
	if len(matches) > 0 {
		var reply string
		comm := matches[1]
		//commArgs := regexp.MustCompile(`\s+`).Split(matches[2], -1)
		switch comm {
		case `help`:
			reply = HelpMessage
		case `start`:
			reply = "Glad to see you, " + UserName + "!"
		}
		if reply != `` {
			msg := tb.NewMessage(ChatID, reply)
			bot.Send(msg)
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
	//readParams(config)
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
			go Talks(config.GetString("Userfile", ""), Bot, update)
		}
	}
}
