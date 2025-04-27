package server

import tele "gopkg.in/telebot.v4"

func MessageFilter(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Chat().Type != tele.ChatPrivate {
			return nil
		}
		return next(c)
	}
}
