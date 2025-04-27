package server

import (
	tele "gopkg.in/telebot.v4"
	"mail_bot/data"
	"mail_bot/mail"
	"mail_bot/utils"
)

type AddMailState struct {
	Email  *string
	Config data.MailConfig
}

var addMailStates map[int64]*AddMailState

func init() {
	RegisterPlainTextHandler("add_mail", handlePlainText)
}

func HandleAddEmail(c tele.Context) error {
	uid := c.Sender().ID
	mails, err := data.GetMail(uid)
	if err != nil {
		err := c.Send("Failed to retrieve email accounts.")
		utils.LogError(err)
		return err
	}
	if len(mails) > MaxUserEmailCount {
		err := c.Send("You have too many email accounts. Please delete some before adding new ones.")
		utils.LogError(err)
		return err
	}
	totalMails, err := data.CountTotalMails()
	if err != nil {
		err := c.Send("Failed to retrieve email accounts.")
		utils.LogError(err)
		return err
	}
	if totalMails > int(MaxEmailCount) {
		err := c.Send("The email server is full. Please try again later.")
		utils.LogError(err)
		return err
	}
	if addMailStates == nil {
		addMailStates = make(map[int64]*AddMailState)
	}
	addMailStates[uid] = &AddMailState{}
	err = c.Send("Please enter your email address. You can cancel the process by typing /cancel.")
	if err != nil {
		utils.LogError(err)
		return err
	}
	return nil
}

func CancelAddEmail(c tele.Context) error {
	uid := c.Sender().ID
	if addMailStates != nil {
		delete(addMailStates, uid)
	}
	err := c.Send("Email addition canceled.")
	if err != nil {
		utils.LogError(err)
		return err
	}
	return nil
}

func handlePlainText(c tele.Context) bool {
	state, isAdding := addMailStates[c.Sender().ID]
	if !isAdding {
		return false
	}

	text := c.Text()

	if state.Email == nil {
		if !utils.IsValidEmail(text) {
			err := c.Send("Please enter a valid email address.")
			utils.LogError(err)
			return true
		} else {
			if data.ExistsMail(c.Sender().ID, text) {
				err := c.Send("This email is already added.")
				utils.LogError(err)
				return true
			}
			state.Email = &text
			state.Config = &data.ImapConfig{} // Currently only IMAP is supported
			err := c.Send(state.Config.NextPrompt())
			utils.LogError(err)
		}
	} else {
		err := state.Config.HandleInput(text)
		if err != nil {
			err := c.Send(err.Error())
			utils.LogError(err)
			return true
		}
		if state.Config.IsFinished() {
			err := c.Send("Configuration complete.")
			utils.LogError(err)
			conn := mail.NewImapConnection(c.Sender().ID, *state.Email, state.Config.(*data.ImapConfig))
			err = mail.AddConnection(conn)
			if err == nil {
				conn.AddListener(HandleNewEmail)
				data.AddMail(c.Sender().ID, *state.Email, state.Config)
				err := c.Send("Email added successfully.")
				utils.LogError(err)
			} else {
				err := c.Send("Failed to add email. Please check your configuration.")
				utils.LogError(err)
			}
			delete(addMailStates, c.Sender().ID)
		} else {
			err := c.Send(state.Config.NextPrompt())
			utils.LogError(err)
		}
	}

	return true
}
