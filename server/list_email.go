package server

import (
	tele "gopkg.in/telebot.v4"
	"mail_bot/data"
)

func HandleListEmails(c tele.Context) error {
	emails, err := data.GetMail(c.Sender().ID)
	if err != nil {
		return c.Send("Error retrieving emails: " + err.Error())
	}
	if len(emails) == 0 {
		return c.Send("No emails found.")
	}
	res := "Your emails:\n"
	for _, email := range emails {
		res += email.Email + "\n"
	}
	return c.Send(res)
}
