package data

import (
	"encoding/json"
	"fmt"
)

type MailConfig interface {
	MailType() string

	ToString() string

	FromString(string) error

	NextPrompt() string

	HandleInput(string) error

	IsFinished() bool
}

type ImapConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

func (c *ImapConfig) MailType() string {
	return "imap"
}

func (c *ImapConfig) ToString() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func (c *ImapConfig) FromString(data string) error {
	return json.Unmarshal([]byte(data), c)
}

func (c *ImapConfig) NextPrompt() string {
	if c.Host == "" {
		return "Please enter the IMAP server host:"
	} else if c.Port == 0 {
		return "Please enter the IMAP server port:"
	} else if c.Username == "" {
		return "Please enter the email username: \n(usually the email address)"
	} else if c.Password == "" {
		return "Please enter the email password:"
	}
	return "Configuration complete."
}

func (c *ImapConfig) HandleInput(input string) error {
	if c.Host == "" {
		c.Host = input
	} else if c.Port == 0 {
		var port int
		_, err := fmt.Sscanf(input, "%d", &port)
		if err != nil {
			return err
		}
		c.Port = port
	} else if c.Username == "" {
		c.Username = input
	} else if c.Password == "" {
		c.Password = input
	}
	return nil
}

func (c *ImapConfig) IsFinished() bool {
	return c.Host != "" && c.Port != 0 && c.Username != "" && c.Password != ""
}

type UserEmailWithConfig struct {
	UserId int64
	Email  string
	Config MailConfig
}
