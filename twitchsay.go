package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"

	"github.com/nickvanw/ircx"
	"github.com/sorcix/irc"
)

var (
	Version string

	user = flag.String("user", "", "Twitch username")
	pass = flag.String("pass", "", "Twitch OAuth token")
	rate = flag.String("rate", "300", "TTS speech rate")

	server = "irc.twitch.tv:6667"
	queue  = make(chan string)
)

func init() {
	flag.Parse()
	go func() {
		for text := range queue {
			Say(text)
		}
	}()
}

func main() {
	bot := ircx.WithLogin(server, *user, *user, *pass)
	if err := bot.Connect(); err != nil {
		log.Panicln("Unable to dial IRC Server ", err)
	}
	RegisterHandlers(bot)
	bot.HandleLoop()
	log.Println("Exiting...")
}

func RegisterHandlers(bot *ircx.Bot) {
	bot.HandleFunc(irc.RPL_WELCOME, RegisterConnect)
	bot.HandleFunc(irc.PING, PingHandler)
	bot.HandleFunc(irc.PRIVMSG, MsgHandler)
	bot.HandleFunc(irc.JOIN, JoinHandler)
	bot.HandleFunc(irc.PART, PartHandler)
}

func RegisterConnect(s ircx.Sender, m *irc.Message) {
	channel := fmt.Sprintf("#%s", *user)
	fmt.Println("Connected, joining", channel, "...")
	s.Send(&irc.Message{
		Command: irc.JOIN,
		Params:  []string{channel},
	})
}

func JoinHandler(s ircx.Sender, m *irc.Message) {
	queue <- fmt.Sprintf("%s has joined.", m.Prefix.Name)
}

func PartHandler(s ircx.Sender, m *irc.Message) {
	queue <- fmt.Sprintf("%s has left.", m.Prefix.Name)
}

func MsgHandler(s ircx.Sender, m *irc.Message) {
	queue <- m.Prefix.Name + ": " + m.Trailing
}

func Say(text string) {
	path, err := exec.LookPath("say")
	if err != nil {
		log.Fatal("can't find say")
	}
	fmt.Println(text)
	cmd := exec.Command(path, "-r", *rate, text)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func PingHandler(s ircx.Sender, m *irc.Message) {
	s.Send(&irc.Message{
		Command:  irc.PONG,
		Params:   m.Params,
		Trailing: m.Trailing,
	})
}
