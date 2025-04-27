package server

import (
	tele "gopkg.in/telebot.v4"
	"mail_bot/data"
)

func RecordUserMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		// Record the user ID and username
		userID := c.Sender().ID
		chatID := c.Chat().ID

		data.AddUser(userID, chatID)

		return next(c)
	}
}
