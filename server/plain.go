package server

import tele "gopkg.in/telebot.v4"

type PlainTextHandler func(c tele.Context) bool

var plainTextHandlers map[string]PlainTextHandler

func HandlePlainText(c tele.Context) error {
	for _, handler := range plainTextHandlers {
		if handler(c) {
			return nil
		}
	}
	return nil
}

func RegisterPlainTextHandler(name string, handler PlainTextHandler) {
	if plainTextHandlers == nil {
		plainTextHandlers = make(map[string]PlainTextHandler)
	}
	plainTextHandlers[name] = handler
}

func UnregisterPlainTextHandler(name string) {
	if plainTextHandlers == nil {
		return
	}
	delete(plainTextHandlers, name)
}
