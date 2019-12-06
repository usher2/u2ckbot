//go:generate protoc -I msg --go_out=plugins=grpc:msg msg/msg.proto
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	tb "github.com/go-telegram-bot-api/telegram-bot-api"

	pb "github.com/usher2/u2ckbot/msg"
	"google.golang.org/grpc"
)

type TypeConfig struct {
	// Config
	Token         string
	UserFile      string
	CkDumpService string
}

// connect to Telegram API
func GetBot(token, loglevel string) *tb.BotAPI {
	bot, err := tb.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	if loglevel == "Debug" {
		bot.Debug = true
	}
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

func main() {
	confFilename := flag.String("c", "u2ckbot.conf", "Configuration file")
	flag.Parse()
	config, err := ReadConfigFile(*confFilename)
	if err != nil {
		log.Fatal(err)
	}
	logLevel := config.GetString("LogLevel", "Debug")
	switch logLevel {
	case "Info":
		logInit(ioutil.Discard, os.Stdout, os.Stderr, os.Stderr)
	case "Warning":
		logInit(ioutil.Discard, ioutil.Discard, os.Stderr, os.Stderr)
	case "Error":
		logInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	default:
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
	bot := GetBot(config.GetString("Token", ""), logLevel)
	// init Users cache
	//ReadUsers(config.UserFile)
	// init update chan
	updates := GetUpdatesChan(bot)
	// read updates
	botUpdates(c, bot, updates)
}
