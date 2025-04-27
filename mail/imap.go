package mail

import (
	"fmt"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"io"
	"log"
	"mail_bot/data"
	"mime"
	"strings"
	"time"
)

type ImapConnection struct {
	userId              int64
	email               string
	config              *data.ImapConfig
	status              Status
	idleCommand         *imapclient.IdleCommand
	listeners           []Listener
	client              *imapclient.Client
	handlingNewMessages bool
	mailIndex           *uint32
}

func (i *ImapConnection) ErrorAndClose() {
	i.status = StatusError
	_ = i.client.Close()
}

func (i *ImapConnection) HandleNewMails(data *imapclient.UnilateralDataMailbox) {
	if data.NumMessages != nil {
		i.handlingNewMessages = true
		defer func() {
			i.handlingNewMessages = false
		}()
		err := i.idleCommand.Close()
		if err != nil {
			log.Println("failed to close idle command:", err)
			i.ErrorAndClose()
			return
		}
		err = i.idleCommand.Wait()
		if err != nil {
			log.Println("failed to close idle command:", err)
			i.ErrorAndClose()
			return
		}
		i.idleCommand = nil

		maxSeqNum, err := getMaxSeqNum(i.client)
		if err != nil {
			log.Println("failed to get max seq num:", err)
			i.ErrorAndClose()
			return
		}

		toBeFetched := make([]int, 0)
		for i := maxSeqNum - *data.NumMessages + 1; i <= maxSeqNum; i++ {
			toBeFetched = append(toBeFetched, int(i))
		}
		messages, err := FetchMessages(i.client, toBeFetched)
		if err != nil {
			log.Println("failed to fetch messages:", err)
			i.ErrorAndClose()
			return
		}
		for _, message := range messages {
			for _, listener := range i.listeners {
				listener(i.userId, message.From, message.To, message.Subject, message.Content)
			}
		}

		i.idleCommand, err = i.client.Idle()
		if err != nil {
			log.Println("failed to start idle command:", err)
			i.ErrorAndClose()
		}
		go func() {
			err := i.idleCommand.Wait()
			if !i.handlingNewMessages {
				log.Println("idle command closed by server:", err)
				i.idleCommand = nil
				i.ErrorAndClose()
			}
		}()
	}
}

func (i *ImapConnection) StartIdle() {
	var err error
	i.idleCommand, err = i.client.Idle()
	if err != nil {
		i.idleCommand = nil
		i.ErrorAndClose()
		log.Println("failed to start idle command: %w", err)
	}
	go func() {
		_ = i.idleCommand.Wait()
		if !i.handlingNewMessages {
			// Idle command closed by server
			_ = i.client.Close()
			_ = i.Connect()
		}
	}()
}

func (i *ImapConnection) StartLoop() {
	go func() {
		for {
			if i.Status() != StatusConnected {
				return
			}
			maxSeqNum, err := getMaxSeqNum(i.client)
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					// Connection closed, reconnect
					_ = i.Connect()
					break
				}
				log.Println("failed to get max seq num:", err)
				i.ErrorAndClose()
				return
			}
			if i.mailIndex == nil {
				i.mailIndex = new(uint32)
				*i.mailIndex = maxSeqNum
			} else if *i.mailIndex < maxSeqNum {
				toBeFetched := make([]int, 0)
				for num := *i.mailIndex + 1; num <= maxSeqNum; num++ {
					toBeFetched = append(toBeFetched, int(num))
				}
				messages, err := FetchMessages(i.client, toBeFetched)
				if err != nil {
					log.Println("failed to fetch messages:", err)
					i.ErrorAndClose()
					return
				}
				for _, message := range messages {
					for _, listener := range i.listeners {
						listener(i.userId, message.From, message.To, message.Subject, message.Content)
					}
				}
				*i.mailIndex = maxSeqNum
			} else if *i.mailIndex > maxSeqNum {
				// This can happen if the server deletes messages
				// Update mailIndex
				*i.mailIndex = maxSeqNum
			}
			time.Sleep(time.Minute)
		}
	}()
}

func (i *ImapConnection) Connect() error {
	i.status = StatusConnecting
	config := i.config
	var err error
	i.client, err = imapclient.DialTLS(fmt.Sprintf("%s:%d", config.Host, config.Port), &imapclient.Options{
		UnilateralDataHandler: &imapclient.UnilateralDataHandler{
			Mailbox: i.HandleNewMails,
		},
		WordDecoder: &mime.WordDecoder{CharsetReader: charset.Reader},
	})
	if err != nil {
		i.status = StatusError
		return fmt.Errorf("failed to connect to IMAP server: %w", err)
	}
	loginCommand := i.client.Login(config.Username, config.Password)
	err = loginCommand.Wait()
	if err != nil {
		i.status = StatusError
		return fmt.Errorf("failed to login: %w", err)
	}
	selectCommand := i.client.Select("INBOX", nil)
	if _, err = selectCommand.Wait(); err != nil {
		i.status = StatusError
		return fmt.Errorf("failed to select mailbox: %w", err)
	}
	idle, err := supportIdle(i.client)
	if err != nil {
		i.status = StatusError
		return fmt.Errorf("failed to check IDLE support: %w", err)
	}
	i.status = StatusConnected
	if idle {
		i.StartIdle()
	} else {
		i.StartLoop()
	}
	return nil
}

func (i *ImapConnection) Disconnect() {
	if i.idleCommand != nil {
		err := i.idleCommand.Close()
		if err != nil {
			log.Println("failed to close idle command:", err)
		}
		i.idleCommand = nil
	}
	_ = i.client.Close()
	i.client = nil
	i.status = StatusDisconnected
}

func (i *ImapConnection) Status() Status {
	return i.status
}

func (i *ImapConnection) GetUserID() int64 {
	return i.userId
}

func (i *ImapConnection) GetEmail() string {
	return i.email
}

func NewImapConnection(userId int64, email string, config *data.ImapConfig) *ImapConnection {
	return &ImapConnection{
		userId: userId,
		email:  email,
		config: config,
		status: StatusDisconnected,
	}
}

func (i *ImapConnection) AddListener(listener Listener) {
	i.listeners = append(i.listeners, listener)
}

func supportIdle(c *imapclient.Client) (bool, error) {
	// Check if the server supports IDLE
	if c == nil {
		return false, nil
	}
	command := c.Capability()
	caps, err := command.Wait()
	if err != nil {
		return false, err
	}
	return caps.Has(imap.CapIdle), nil
}

type Message struct {
	ID      uint32
	From    string
	To      string
	Subject string
	Content string
}

func FetchMessages(c *imapclient.Client, seqNums []int) ([]Message, error) {
	bodySection := &imap.FetchItemBodySection{}
	s := imap.SeqSet{}
	for _, seqNum := range seqNums {
		s.AddNum(uint32(seqNum))
	}
	fetchCommand := c.Fetch(s, &imap.FetchOptions{
		BodySection: []*imap.FetchItemBodySection{bodySection},
	})
	var messages []Message
	for message := fetchCommand.Next(); message != nil; message = fetchCommand.Next() {
		var bodySectionData imapclient.FetchItemDataBodySection
		ok := false
		for {
			item := message.Next()
			if item == nil {
				break
			}

			switch v := item.(type) {
			case imapclient.FetchItemDataBodySection:
				bodySectionData = v
				ok = true
			}

			if ok {
				break
			}
		}
		if !ok {
			continue
		}
		mr, err := mail.CreateReader(bodySectionData.Literal)
		if err != nil {
			return nil, err
		}
		h := mr.Header
		fromList, err := h.AddressList("From")
		from := "Unknown"
		if len(fromList) != 0 {
			from = fromList[0].String()
		}
		toList, err := h.AddressList("To")
		to := "Unknown"
		if len(toList) != 0 {
			to = toList[0].String()
		}
		subject, err := h.Subject()
		if err != nil {
			subject = "Unknown"
		}
		content := ""
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			switch p.Header.(type) {
			case *mail.InlineHeader:
				d, _ := io.ReadAll(p.Body)
				if len(d) > 0 {
					content = string(d)
				}
			}

			if len([]rune(content)) > 1000 {
				content = string([]rune(content)[:1000])
				break
			}
		}
		messages = append(messages, Message{
			ID:      message.SeqNum,
			From:    from,
			To:      to,
			Subject: subject,
			Content: content,
		})
	}
	_ = fetchCommand.Close()
	return messages, nil
}

func getMaxSeqNum(c *imapclient.Client) (uint32, error) {
	// Get the highest message sequence number
	selectCommand := c.Select("INBOX", nil)
	d, err := selectCommand.Wait()
	if err != nil {
		return 0, fmt.Errorf("failed to select mailbox: %w", err)
	}
	return d.NumMessages, nil
}
