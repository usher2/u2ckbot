//go:generate  protoc -I msg --go-grpc_out=msg --go_out=msg --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative msg/msg.proto
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	tb "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/usher2/u2ckbot/internal/logger"
	pb "github.com/usher2/u2ckbot/msg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const MAXMSGSIZE = 1024 * 1024 * 128

type TypeConfig struct {
	// Config
	Token         string
	UserFile      string
	CkDumpService string
	HTTPSProxyUrl string
}

// connect to Telegram API
func GetBot(token, proxyUrl, loglevel string) *tb.BotAPI {
	var bot *tb.BotAPI
	var err error
	if proxyUrl != "" {
		var _proxyUrl *url.URL
		_proxyUrl, err = url.Parse(proxyUrl)
		if err != nil {
			log.Panic("Proxy url invalid")
		}
		client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(_proxyUrl)}}
		bot, err = tb.NewBotAPIWithClient(token, tb.APIEndpoint, client)
	} else {
		bot, err = tb.NewBotAPI(token)
	}
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
	updates := bot.GetUpdatesChan(c)

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
		logger.LogInit(io.Discard, os.Stdout, os.Stderr, os.Stderr)
	case "Warning":
		logger.LogInit(io.Discard, io.Discard, os.Stderr, os.Stderr)
	case "Error":
		logger.LogInit(io.Discard, io.Discard, io.Discard, os.Stderr)
	default:
		logger.LogInit(os.Stderr, os.Stdout, os.Stderr, os.Stderr)
	}
	// gRPC
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MAXMSGSIZE)))
	// opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(config.GetString("CkDumpServer", "localhost:50001"), opts...)
	if err != nil {
		fmt.Printf("fail to dial: %v", err)
	}
	defer conn.Close()
	fmt.Printf("Connect...\n")
	c := pb.NewCheckClient(conn)
	// connect to Telegram API
	bot := GetBot(config.GetString("Token", ""), config.GetString("Proxy", ""), logLevel)
	// init update chan
	updates := GetUpdatesChan(bot)
	// read updates
	botUpdates(c, bot, updates)
}
