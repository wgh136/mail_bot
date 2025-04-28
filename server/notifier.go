package server

import (
	"github.com/k3a/html2text"
	tele "gopkg.in/telebot.v4"
	"log"
	"mail_bot/data"
)

var Bot *tele.Bot

func HandleNewEmail(uid int64, from string, to string, subject string, body string) {
	if Bot == nil {
		return
	}
	chatID, err := data.GetUserChatId(uid)
	if err != nil {
		log.Println("Error getting chat ID for user ", uid, ":", err)
	}
	body = html2text.HTML2Text(body)
	if len([]rune(body)) > 2000 {
		body = string([]rune(body)[:2000])
	}
	_, err = Bot.Send(tele.ChatID(chatID), "From: "+from+"\nTo: "+to+"\nSubject: "+subject+"\n\n"+body)
	if err != nil {
		log.Println("Error sending message to user ", uid, ":", err)
	}
}

func SetBot(bot *tele.Bot) {
	Bot = bot
}
