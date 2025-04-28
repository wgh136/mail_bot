package main

import (
	tele "gopkg.in/telebot.v4"
	"log"
	"mail_bot/data"
	"mail_bot/mail"
	"mail_bot/server"
	"os"
	"time"
)

const helloMessage = `Welcome to the Mail Bot! Here is what you can do:
/add_email - Add a new email address.
/delete_email - Delete an existing email address.
/list_email - List all your email addresses.
/cancel - Cancel the current operation.`

func main() {
	token := os.Getenv("TOKEN")
	if token == "" {
		panic("Please set the TOKEN environment variable.")
	}

	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	server.SetBot(bot)

	bot.Use(server.MessageFilter)

	bot.Use(server.RecordUserMiddleware)

	bot.Handle("/start", func(c tele.Context) error {
		return c.Send(helloMessage)
	})

	bot.Handle("/add_email", server.HandleAddEmail)

	bot.Handle("/cancel", server.CancelAddEmail)

	bot.Handle("/delete_email", server.HandleDeleteEmail)

	bot.Handle("/list_email", server.HandleListEmails)

	bot.Handle(&tele.InlineButton{
		Unique: server.InlineButtonDeleteEmail,
	}, server.HandleDeleteEmailButton)

	bot.Handle(tele.OnText, server.HandlePlainText)

	bot.Start()
}

func init() {
	configs := data.GetAllConfigs()
	for _, c := range configs {
		mailConfig := c.Config
		conn := mail.NewImapConnection(c.UserId, c.Email, mailConfig.(*data.ImapConfig))
		conn.AddListener(server.HandleNewEmail)
		err := mail.AddConnection(conn)
		if err != nil {
			log.Println("Failed to add connection:", err)
		}
	}
}
