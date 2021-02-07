package client

import (
	"github.com/curltech/go-colla-core/logger"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type ImapClient struct {
	*client.Client
	addr     string
	username string
	password string
}

var imapClient *ImapClient

func init() {
	imapClient = &ImapClient{}
}

func GetImapClient() *ImapClient {
	return imapClient
}

func (this *ImapClient) Login(addr string, username string, password string) error {
	logger.Sugar.Infof("Connecting to server...")
	this.addr = addr
	this.username = username
	this.password = password
	// Connect to server
	var err error
	this.Client, err = client.DialTLS(this.addr, nil)
	if err != nil {
		logger.Sugar.Errorf(err.Error())

		return err
	}
	logger.Sugar.Infof("Connected")

	// Login
	if err = this.Client.Login(username, password); err != nil {
		logger.Sugar.Errorf(err.Error())

		return err
	}
	logger.Sugar.Infof("Logged in")

	return nil
}

func (this *ImapClient) receive() {
	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- this.Client.List("", "*", mailboxes)
	}()

	logger.Sugar.Infof("Mailboxes:")
	for m := range mailboxes {
		logger.Sugar.Infof("* " + m.Name)
	}

	if err := <-done; err != nil {
		logger.Sugar.Errorf(err.Error())

		return
	}

	// Select INBOX
	mbox, err := this.Client.Select("INBOX", false)
	if err != nil {
		logger.Sugar.Errorf(err.Error())

		return
	}
	logger.Sugar.Infof("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- this.Client.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	logger.Sugar.Infof("Last 4 messages:")
	for msg := range messages {
		logger.Sugar.Infof("* " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		logger.Sugar.Errorf(err.Error())

		return
	}

	logger.Sugar.Infof("Done!")
}

func (this *ImapClient) Logout() {
	// Don't forget to logout
	this.Logout()
}
