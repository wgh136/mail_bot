package server

import (
	"fmt"
	tele "gopkg.in/telebot.v4"
	"mail_bot/data"
	"mail_bot/mail"
	"mail_bot/utils"
)

const (
	InlineButtonDeleteEmail = "DeleteEmail"
)

func HandleDeleteEmail(c tele.Context) error {
	uid := c.Sender().ID
	emails, err := data.GetMail(uid)
	if err != nil {
		utils.LogError(err)
		return err
	}
	if len(emails) == 0 {
		err := c.Send("No emails found. Please add an email first.")
		utils.LogError(err)
		return err
	}
	buttons := make([]tele.InlineButton, len(emails))
	for i, email := range emails {
		buttons[i] = tele.InlineButton{
			Unique: InlineButtonDeleteEmail,
			Text:   email.Email,
			Data:   email.Email,
		}
	}
	err = c.Send("Please select the email you want to delete:", &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			buttons,
		},
	})
	if err != nil {
		utils.LogError(err)
		return err
	}
	return nil
}

func HandleDeleteEmailButton(c tele.Context) error {
	email := c.Data()

	if len(email) == 0 {
		_ = c.Send("No email provided.")
		return nil
	}
	data.DeleteMail(c.Sender().ID, email)
	mail.RemoveConnection(c.Sender().ID, email)
	_ = c.Send(fmt.Sprintf("Email %s deleted successfully.", email))
	_ = c.Respond()
	return nil
}
